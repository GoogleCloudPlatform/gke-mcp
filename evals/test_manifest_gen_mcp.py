import json
import os
import subprocess
import time
from deepeval import assert_test
from deepeval.metrics import GEval
from deepeval.models.base_model import DeepEvalBaseLLM
from deepeval.test_case import LLMTestCase
from deepeval.test_case import LLMTestCaseParams
from langchain_google_genai import ChatGoogleGenerativeAI
import pytest
import yaml


# Custom model implementation for Gemini in DeepEval using Google AI Studio API Key
class GoogleGeminiAI(DeepEvalBaseLLM):

  def __init__(self, model):
    self.model = model

  def load_model(self):
    return self.model

  def generate(self, prompt: str) -> str:
    chat_model = self.load_model()
    return chat_model.invoke(prompt).content

  async def a_generate(self, prompt: str) -> str:
    chat_model = self.load_model()
    res = await chat_model.ainvoke(prompt)
    return res.content

  def get_model_name(self):
    return "Gemini AI Model"


def start_mcp_server():
  """Starts the Go MCP server in stdio mode."""
  return subprocess.Popen(
      ["go", "run", "main.go"],
      stdin=subprocess.PIPE,
      stdout=subprocess.PIPE,
      stderr=subprocess.PIPE,
      text=True,
      bufsize=1,  # Line buffered
  )


def send_rpc_message(process, message):
  """Sends a JSON-RPC message to the process stdin."""
  process.stdin.write(json.dumps(message) + "\n")
  process.stdin.flush()


def read_rpc_response(process):
  """Reads a JSON-RPC response from the process stdout."""
  return process.stdout.readline()


def mcp_initialize(process):
  """Performs the MCP initialization handshake."""
  # 1. Send initialize request
  init_req = {
      "jsonrpc": "2.0",
      "id": 1,
      "method": "initialize",
      "params": {
          "protocolVersion": "2024-11-05",
          "capabilities": {},
          "clientInfo": {"name": "test-client", "version": "1.0.0"},
      },
  }
  send_rpc_message(process, init_req)

  # Read response
  init_res_str = read_rpc_response(process)
  print(f"Init Response: {init_res_str}")

  # 2. Send initialized notification
  initialized_notif = {
      "jsonrpc": "2.0",
      "method": "notifications/initialized",
      "params": {},
  }
  send_rpc_message(process, initialized_notif)


def mcp_call_tool(process, tool_name, arguments):
  """Calls an MCP tool and returns the response."""
  call_req = {
      "jsonrpc": "2.0",
      "id": 2,
      "method": "tools/call",
      "params": {"name": tool_name, "arguments": arguments},
  }
  send_rpc_message(process, call_req)

  # Read response
  call_res_str = read_rpc_response(process)
  print(f"Call Response: {call_res_str}")
  return json.loads(call_res_str)


def test_manifest_generation_mcp():
  prompt = (
      "Generate Kubernetes manifest for deploying our finetuned Gemma 3 12b"
      " model to our GKE cluster"
  )

  server_process = start_mcp_server()

  output = ""
  try:
    mcp_initialize(server_process)

    call_res = mcp_call_tool(
        server_process, "generate_manifest", {"prompt": prompt}
    )

    # Extract output
    content = call_res["result"]["content"]
    for c in content:
      if c["type"] == "text":
        output += c["text"]

  except Exception as e:
    print(f"Error communicating with MCP server: {e}")
    # Read stderr to see what went wrong
    stderr = server_process.stderr.read()
    print(f"Server Stderr: {stderr}")
    pytest.fail(f"Failed to communicate with MCP server: {e}")
  finally:
    server_process.terminate()
    server_process.wait()

  print(f"Raw Generated Output:\n{output}")

  # Cleanup markdown code blocks if present
  cleaned_output = output.strip()
  if cleaned_output.startswith("```yaml"):
    cleaned_output = cleaned_output[len("```yaml") :]
  if cleaned_output.endswith("```"):
    cleaned_output = cleaned_output[: -len("```")]
  cleaned_output = cleaned_output.strip()

  print(f"Cleaned Manifest:\n{cleaned_output}")

  # 1. Hard check: Verify it is valid YAML
  try:
    documents = list(yaml.safe_load_all(cleaned_output))
    print(f"YAML is valid. Found {len(documents)} documents.")
    assert len(documents) > 0, "No YAML documents found"
    for doc in documents:
      if doc:
        assert "apiVersion" in doc, "Missing apiVersion in a document"
        assert "kind" in doc, "Missing kind in a document"
  except yaml.YAMLError as exc:
    pytest.fail(f"Output is not valid YAML: {exc}")
  except AssertionError as exc:
    pytest.fail(f"YAML is valid but not a Kubernetes manifest: {exc}")

  # 2. DeepEval check: Use GEval with Gemini
  api_key = os.getenv("GEMINI_API_KEY")
  if not api_key:
    pytest.fail("GEMINI_API_KEY environment variable not set")

  chat_model = ChatGoogleGenerativeAI(
      model="gemini-2.5-flash", google_api_key=api_key
  )
  gemini_ai_model = GoogleGeminiAI(model=chat_model)

  valid_yaml_metric = GEval(
      name="Valid YAML Manifest",
      criteria=(
          "The output is a valid Kubernetes manifest in YAML format and"
          " addresses the request."
      ),
      evaluation_params=[
          LLMTestCaseParams.INPUT,
          LLMTestCaseParams.ACTUAL_OUTPUT,
      ],
      model=gemini_ai_model,
  )

  test_case = LLMTestCase(input=prompt, actual_output=cleaned_output)

  assert_test(test_case, [valid_yaml_metric])
