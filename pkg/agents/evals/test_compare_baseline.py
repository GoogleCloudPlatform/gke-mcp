import subprocess
import json
import time
import os
import re
import pytest
import yaml
from deepeval import evaluate
from deepeval.test_case import LLMTestCase
from deepeval.metrics import GEval, AnswerRelevancyMetric
from deepeval.test_case import LLMTestCaseParams
from deepeval.models.base_model import DeepEvalBaseLLM
from langchain_google_genai import ChatGoogleGenerativeAI

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
        bufsize=1 # Line buffered
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
    init_req = {
        "jsonrpc": "2.0",
        "id": 1,
        "method": "initialize",
        "params": {
            "protocolVersion": "2024-11-05",
            "capabilities": {},
            "clientInfo": {"name": "test-client", "version": "1.0.0"}
        }
    }
    send_rpc_message(process, init_req)
    read_rpc_response(process)
    
    initialized_notif = {
        "jsonrpc": "2.0",
        "method": "notifications/initialized",
        "params": {}
    }
    send_rpc_message(process, initialized_notif)

def mcp_call_tool(process, tool_name, arguments):
    """Calls an MCP tool and returns the response."""
    call_req = {
        "jsonrpc": "2.0",
        "id": 2,
        "method": "tools/call",
        "params": {
            "name": tool_name,
            "arguments": arguments
        }
    }
    send_rpc_message(process, call_req)
    call_res_str = read_rpc_response(process)
    
    if not call_res_str:
        raise RuntimeError("MCP server terminated unexpectedly or closed stdout")
        
    try:
        response = json.loads(call_res_str)
    except json.JSONDecodeError as e:
        raise RuntimeError(f"Failed to parse JSON-RPC response: {e}. Raw response: {call_res_str!r}")
        
    if "error" in response:
        error_data = response["error"]
        code = error_data.get("code")
        message = error_data.get("message")
        data = error_data.get("data")
        raise RuntimeError(f"MCP tool call failed with error code {code}: {message}. Data: {data}")
        
    return response

def clean_yaml(output):
    """Cleans up markdown code blocks if present."""
    
    # First try to find a block explicitly labeled as yaml
    yaml_pattern = r"\x60\x60\x60yaml\s*(.*?)\s*\x60\x60\x60"
    match = re.search(yaml_pattern, output, re.DOTALL | re.IGNORECASE)
    if match:
        return match.group(1).strip()
        
    # If not found, fall back to finding the first code block
    any_pattern = r"\x60\x60\x60(?:[a-zA-Z]+)?\s*(.*?)\s*\x60\x60\x60"
    match = re.search(any_pattern, output, re.DOTALL)
    if match:
        return match.group(1).strip()
        
    return output.strip()

def generate_markdown_report(results, prompt, agent_output, baseline_output, agent_cleaned):
    """Generates a beautiful markdown report for GitHub Job Summary."""
    
    agent_res = None
    baseline_res = None
    
    for res in results.test_results:
        actual = getattr(res, 'actual_output', None)
        if actual is None:
             test_case = getattr(res, 'test_case', None)
             if test_case:
                  actual = getattr(test_case, 'actual_output', None)
                  
        if actual == agent_cleaned:
            agent_res = res
        else:
            baseline_res = res

    def get_metric(metrics_data, name):
        for m in metrics_data:
            if name in m.name:
                return m
        return None
        
    agent_yaml = get_metric(agent_res.metrics_data, "Valid YAML Manifest")
    agent_rel = get_metric(agent_res.metrics_data, "Answer Relevancy")
    agent_hallucination = get_metric(agent_res.metrics_data, "Hallucination Check")
    
    baseline_yaml = get_metric(baseline_res.metrics_data, "Valid YAML Manifest")
    baseline_rel = get_metric(baseline_res.metrics_data, "Answer Relevancy")
    baseline_hallucination = get_metric(baseline_res.metrics_data, "Hallucination Check")
    
    def format_score(m):
        if not m:
            return "N/A"
        emoji = "✅" if m.success else "❌"
        return f"{emoji} {m.score:.2f}"

    def get_status(m_agent, m_baseline):
        if not m_agent or not m_baseline:
            return "N/A"
        if not m_agent.success and not m_baseline.success:
            return "Both Failed"
        if m_baseline.score > m_agent.score:
            return "Baseline Won"
        if m_agent.score > m_baseline.score:
            return "Agent Won"
        return "Tie"

    markdown = f"""# 🤖 GKE-MCP Baseline Evaluation Report

This report compares the performance of the **Agentic MCP Tool** against the **Baseline Gemini Model** on the same prompt.

> [!IMPORTANT]
> **Prompt**: {prompt}

## 🎯 Metrics Summary

| Metric | Agentic MCP Tool | Baseline Gemini | Status |
| :--- | :---: | :---: | :--- |
| **Valid YAML Manifest** | {format_score(agent_yaml)} | {format_score(baseline_yaml)} | {get_status(agent_yaml, baseline_yaml)} |
| **Answer Relevancy** | {format_score(agent_rel)} | {format_score(baseline_rel)} | {get_status(agent_rel, baseline_rel)} |
| **Hallucination Check** | {format_score(agent_hallucination)} | {format_score(baseline_hallucination)} | {get_status(agent_hallucination, baseline_hallucination)} |

<details>
<summary>🔍 Click to expand Detailed Analysis</summary>

### 🤖 Agentic MCP Tool
- **Valid YAML Manifest**: {agent_yaml.reason if agent_yaml else "N/A"}
- **Answer Relevancy**: {agent_rel.reason if agent_rel else "N/A"}
- **Hallucination Check**: {agent_hallucination.reason if agent_hallucination else "N/A"}

### 🪐 Baseline Gemini
- **Valid YAML Manifest**: {baseline_yaml.reason if baseline_yaml else "N/A"}
- **Answer Relevancy**: {baseline_rel.reason if baseline_rel else "N/A"}
- **Hallucination Check**: {baseline_hallucination.reason if baseline_hallucination else "N/A"}

</details>

<details>
<summary>📄 Click to expand Raw Outputs</summary>

### 🤖 Agentic MCP Tool
````yaml
{agent_output}
````

### 🪐 Baseline Gemini
````yaml
{baseline_output}
````

</details>

> [!NOTE]
> This report was automatically generated by the benchmark test suite.
"""
    
    report_path = os.path.join(os.path.dirname(__file__), "JOB_SUMMARY.md")
    with open(report_path, "w") as f:
        f.write(markdown)
    print("Generated pkg/agents/evals/JOB_SUMMARY.md")

def test_compare_agent_vs_baseline():
  prompt = "Generate a manifest for model google/gemma-2-2b-it, using vllm server on nvidia-l4 accelerator. Do not use giq_generate_manifest"

  # 1. Get Agent Output
  server_process = start_mcp_server()
  agent_output = ""
  try:
    mcp_initialize(server_process)
    call_res = mcp_call_tool(server_process, "generate_manifest", {"prompt": prompt})
    content = call_res["result"]["content"]
    for c in content:
      if c["type"] == "text":
        agent_output += c["text"]
    print("\n=== Raw Agent Output ===")
    print(agent_output)
    print("========================")
  except Exception as e:
    pytest.fail(f"Failed to communicate with MCP server: {e}")
  finally:
    server_process.terminate()
    server_process.wait()

  agent_cleaned = clean_yaml(agent_output)

  # 2. Get Baseline Output
  # Switch to Vertex AI using ADC set up by the workflow
  project_id = os.getenv("GOOGLE_CLOUD_PROJECT", "gke-mcp-ci-cd")
  
  chat_model = ChatGoogleGenerativeAI(
      model="gemini-2.5-pro",
      vertexai=True,
      project=project_id
  )
  baseline_output = chat_model.invoke(prompt).content
  baseline_cleaned = clean_yaml(baseline_output)

  # 3. Evaluate both
  gemini_ai_model = GoogleGeminiAI(model=chat_model)

  valid_yaml_metric = GEval(
      name="Valid YAML Manifest",
      criteria="The output is a valid Kubernetes manifest in YAML format and addresses the request.",
      evaluation_params=[LLMTestCaseParams.INPUT, LLMTestCaseParams.ACTUAL_OUTPUT],
      model=gemini_ai_model,
      threshold=0.7
  )

  relevance_metric = AnswerRelevancyMetric(
      threshold=0.5,
      model=gemini_ai_model,
      include_reason=True
  )

  hallucination_geval_metric = GEval(
      name="Hallucination Check",
      criteria="The output should not invent or hallucinate non-existent model names, accelerator types, or GKE features. It should stick to the facts provided in the prompt or standard GKE documentation.",
      evaluation_params=[LLMTestCaseParams.INPUT, LLMTestCaseParams.ACTUAL_OUTPUT],
      model=gemini_ai_model
  )

  agent_test_case = LLMTestCase(
        input=prompt,
        actual_output=agent_cleaned
    )

  baseline_test_case = LLMTestCase(
        input=prompt,
        actual_output=baseline_cleaned
    )

  # Run evaluation
  results = evaluate(
      [agent_test_case, baseline_test_case],
      [valid_yaml_metric, relevance_metric, hallucination_geval_metric],
  )

  print(f"Results: {results}")

  # Generate report
  generate_markdown_report(results, prompt, agent_output, baseline_output, agent_cleaned)

  # Check for failures on the Agent to provide pass/fail status
  agent_res = None
  for res in results.test_results:
      actual = getattr(res, 'actual_output', None)
      if actual is None:
           test_case = getattr(res, 'test_case', None)
           if test_case:
                actual = getattr(test_case, 'actual_output', None)
      if actual == agent_cleaned:
          agent_res = res
          break

  if agent_res:
      for metric in agent_res.metrics_data:
        if not metric.success:
          pytest.fail(
              f"Agent Metric '{metric.name}' failed with score {metric.score:.2f}."
              f" Reason: {metric.reason}"
          )
  else:
      pytest.fail("Could not find Agent result in evaluation results")

if __name__ == '__main__':
    test_compare_agent_vs_baseline()
