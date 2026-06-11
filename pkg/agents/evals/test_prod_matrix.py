# Copyright 2026 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

import subprocess
import json
import time
import os
import pytest
import yaml
import requests
from deepeval import evaluate
from deepeval.test_case import LLMTestCase
from deepeval.metrics import GEval, AnswerRelevancyMetric
from deepeval.test_case import LLMTestCaseParams
from deepeval.models.base_model import DeepEvalBaseLLM
from langchain_google_genai import ChatGoogleGenerativeAI
import google.auth
import google.auth.transport.requests

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
    """Starts the MCP server in stdio mode."""
    cmd_str = os.getenv("GKE_MCP_SERVER_CMD", "go run main.go")
    cmd = cmd_str.split()
    return subprocess.Popen(
        cmd,
        stdin=subprocess.PIPE,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        text=True,
        bufsize=1
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
    """Calls an MCP tool over stdio."""
    req = {
        "jsonrpc": "2.0",
        "method": "tools/call",
        "id": 2,
        "params": {
            "name": tool_name,
            "arguments": arguments
        }
    }
    send_rpc_message(process, req)
    
    # Read response lines until we find the JSON-RPC result
    while True:
        line = read_rpc_response(process)
        if not line:
            raise Exception("MCP server closed connection")
        print(f"MCP Server Output: {line.strip()}")
        try:
            res = json.loads(line)
            if "id" in res and res["id"] == 2:
                return res
        except json.JSONDecodeError:
            continue

def clean_yaml(text):
    """Extracts YAML from markdown code blocks."""
    if "```yaml" in text:
        parts = text.split("```yaml")
        return parts[1].split("```")[0].strip()
    elif "```" in text:
        parts = text.split("```")
        return parts[1].split("```")[0].strip()
    return text.strip()

def get_claude_baseline(prompt, model_name="claude-3-7-sonnet-20250219"):
    """Calls Claude on Anthropic API via REST to get baseline."""
    url = "https://api.anthropic.com/v1/messages"
    
    headers = {
        "x-api-key": os.environ.get("ANTHROPIC_API_KEY"),
        "anthropic-version": "2023-06-01",
        "Content-Type": "application/json"
    }
    
    payload = {
        "model": model_name,
        "max_tokens": 1024,
        "messages": [
            {"role": "user", "content": prompt}
        ]
    }
    
    resp = requests.post(url, headers=headers, json=payload)
    if resp.status_code != 200:
        raise Exception(f"Anthropic API error: {resp.status_code} - {resp.text}")
        
    result = resp.json()
    return result["content"][0]["text"]

def generate_markdown_report(results, prompt, display_name):
    """Generates a comprehensive markdown report comparing all subjects."""
    agent_res = results.test_results[0]
    baseline_res = results.test_results[1]
    
    def get_metric_data(res, metric_name):
        for m in res.metrics_data:
            if metric_name in m.name:
                return f"{m.score:.2f}", m.reason if hasattr(m, 'reason') else "N/A"
        return "N/A", "N/A"
        
    agent_prod_score, agent_prod_reason = get_metric_data(agent_res, "Production Readiness")
    agent_yaml_score, agent_yaml_reason = get_metric_data(agent_res, "Valid YAML")
    agent_hall_score, agent_hall_reason = get_metric_data(agent_res, "Hallucination Check")
    agent_rel_score, agent_rel_reason = get_metric_data(agent_res, "Answer Relevancy")
    
    baseline_prod_score, baseline_prod_reason = get_metric_data(baseline_res, "Production Readiness")
    baseline_yaml_score, baseline_yaml_reason = get_metric_data(baseline_res, "Valid YAML")
    baseline_hall_score, baseline_hall_reason = get_metric_data(baseline_res, "Hallucination Check")
    baseline_rel_score, baseline_rel_reason = get_metric_data(baseline_res, "Answer Relevancy")
    
    markdown = f"""# 📊 DeepEval Production Readiness Report for {display_name}

## 📝 Prompt
> {prompt}

## 🎯 Metrics Summary

| Subject | Production Readiness (0-5) | Valid YAML (0-1) | Hallucination Check (0-1) | Answer Relevancy (0-1) |
| :--- | :---: | :---: | :---: | :---: |
| **{display_name} Agent** | {agent_prod_score} | {agent_yaml_score} | {agent_hall_score} | {agent_rel_score} |
| **{display_name} Baseline** | {baseline_prod_score} | {baseline_yaml_score} | {baseline_hall_score} | {baseline_rel_score} |

## 🔍 Detailed Analysis

### 🤖 {display_name} Agent
- **Production Readiness Reason**: {agent_prod_reason}
- **Valid YAML Reason**: {agent_yaml_reason}
- **Hallucination Check Reason**: {agent_hall_reason}
- **Answer Relevancy Reason**: {agent_rel_reason}

### 🪐 {display_name} Baseline
- **Production Readiness Reason**: {baseline_prod_reason}
- **Valid YAML Reason**: {baseline_yaml_reason}
- **Hallucination Check Reason**: {baseline_hall_reason}
- **Answer Relevancy Reason**: {baseline_rel_reason}
"""
    
    filename = f"pkg/agents/evals/PROD_SUMMARY_{display_name.upper()}.md"
    with open(filename, "w") as f:
        f.write(markdown)
    print(f"Generated {filename}")

@pytest.mark.parametrize("provider,model,display_name", [
    ("vertex-ai", "gemini-2.5-flash-lite", "Gemini"),
    ("anthropic", "claude-opus-4-7", "Claude")
])
def test_prod_matrix(provider, model, display_name):
    prompt = "Generate a manifest for model google/gemma-2-2b-it, using vllm server on nvidia-l4 accelerator."
    
    api_key = os.environ.get("GEMINI_API_KEY")
    project_id = os.environ.get("GCP_PROJECT_ID", "jayantid-gke-dev")
    
    # 1. Get Agent Output
    os.environ["GKE_MCP_PROVIDER"] = provider
    os.environ["GKE_MCP_MODEL"] = model
    os.environ["ANTHROPIC_API_KEY"] = os.environ.get("ANTHROPIC_API_KEY", "dummy")
    
    server_process = start_mcp_server()
    agent_cleaned = ""
    try:
        # Perform MCP initialization
        mcp_initialize(server_process)
        
        call_res = mcp_call_tool(server_process, "generate_manifest", {"prompt": prompt})
        content = call_res["result"]["content"]
        for part in content:
            if part["type"] == "text":
                agent_cleaned = clean_yaml(part["text"])
    finally:
        server_process.terminate()
        server_process.wait()
        
    # 2. Get Baseline Output
    baseline_cleaned = ""
    if provider == "vertex-ai":
        chat_model = ChatGoogleGenerativeAI(model=model, google_api_key=api_key)
        baseline_output = chat_model.invoke(prompt).content
        baseline_cleaned = clean_yaml(baseline_output)
    elif provider == "anthropic":
        baseline_output = get_claude_baseline(prompt, model_name=model)
        baseline_cleaned = clean_yaml(baseline_output)

    # 3. Evaluate
    chat_model_judge = ChatGoogleGenerativeAI(model="gemini-2.5-flash-lite", google_api_key=api_key)
    gemini_ai_model = GoogleGeminiAI(model=chat_model_judge)
    
    prod_readiness_metric = GEval(
        name="Production Readiness",
        criteria="""
        Evaluate the generated Kubernetes manifest on a scale of 0 to 5 based on its readiness for production deployment on GKE.
        
        Criteria:
        1. Security Hardening (0-1 points): Does the manifest include security best practices like runAsNonRoot, seccompProfile, dropping capabilities, and readOnlyRootFilesystem?
        2. Reliability & Availability (0-1 points): Are proper liveness, readiness, and startup probes defined with appropriate thresholds?
        3. Resource Management (0-1 points): Are both requests and limits explicitly set for CPU and Memory?
        4. GKE & Hardware Optimization (0-1 points): For AI workloads (e.g. using GPUs/TPUs), are nodeSelector and tolerations used correctly?
        5. Observability & Conventions (0-1 points): Are standard labeling conventions applied (e.g., app.kubernetes.io/name)?
        
        Scoring Rubric:
        - Give 1 point for each criteria met satisfactorily.
        - Total score should be the sum of points, between 0 and 5.
        """,
        evaluation_params=[LLMTestCaseParams.INPUT, LLMTestCaseParams.ACTUAL_OUTPUT],
        model=gemini_ai_model
    )
    
    valid_yaml_metric = GEval(
        name="Valid YAML Manifest",
        criteria="The output is a valid Kubernetes manifest in YAML format and addresses the request.",
        evaluation_params=[LLMTestCaseParams.INPUT, LLMTestCaseParams.ACTUAL_OUTPUT],
        model=gemini_ai_model
    )
    
    relevance_metric = AnswerRelevancyMetric(
        threshold=0.5,
        model=gemini_ai_model,
        include_reason=True
    )
    
    hallucination_geval_metric = GEval(
        name="Hallucination Check",
        criteria="The output should not invent or hallucinate non-existent model names, accelerator types, or GKE features.",
        evaluation_params=[LLMTestCaseParams.INPUT, LLMTestCaseParams.ACTUAL_OUTPUT],
        model=gemini_ai_model
    )
    
    # We use a composite input to distinguish between runs in the final report
    agent_test_case = LLMTestCase(
        input=f"{display_name} Agent | {prompt}",
        actual_output=agent_cleaned
    )
    
    baseline_test_case = LLMTestCase(
        input=f"{display_name} Baseline | {prompt}",
        actual_output=baseline_cleaned
    )
    
    # Run evaluation
    results = evaluate(
        [agent_test_case, baseline_test_case],
        [prod_readiness_metric, valid_yaml_metric, relevance_metric, hallucination_geval_metric],
    )
    
    # Save results for reporting
    generate_markdown_report(results, prompt, display_name)
