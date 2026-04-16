import subprocess
import yaml
import pytest
import os
from deepeval import assert_test
from deepeval.test_case import LLMTestCase
from deepeval.metrics import GEval
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

def test_manifest_generation():
    prompt = "Help me generate a Kubernetes manifest for an Nginx deployment."
    
    # Path to the Go CLI wrapper relative to the repo root
    cli_path = "./cmd/manifestgen_cli/manifestgen_cli"
    
    try:
        # Running the CLI from the repo root
        result = subprocess.run([cli_path, prompt], capture_output=True, text=True, check=True)
        output = result.stdout
    except subprocess.CalledProcessError as e:
        print(f"CLI failed with error: {e.stderr}")
        pytest.fail(f"CLI failed: {e}")

    print(f"Raw Generated Output:\n{output}")

    # Cleanup markdown code blocks if present
    cleaned_output = output.strip()
    if cleaned_output.startswith("```yaml"):
        cleaned_output = cleaned_output[len("```yaml"):]
    if cleaned_output.endswith("```"):
        cleaned_output = cleaned_output[:-len("```")]
    cleaned_output = cleaned_output.strip()

    print(f"Cleaned Manifest:\n{cleaned_output}")

    # 1. Hard check: Verify it is valid YAML (handling multiple documents)
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

    # 2. DeepEval check: Use GEval with Gemini and API key from environment
    
    api_key = os.getenv("GEMINI_API_KEY")
    if not api_key:
        pytest.fail("GEMINI_API_KEY environment variable not set")

    # Using gemini-2.5-flash as listed in the available models
    chat_model = ChatGoogleGenerativeAI(model="gemini-2.5-flash", google_api_key=api_key)
    gemini_ai_model = GoogleGeminiAI(model=chat_model)

    valid_yaml_metric = GEval(
        name="Valid YAML Manifest",
        criteria="The output is a valid Kubernetes manifest in YAML format and addresses the request.",
        evaluation_params=[LLMTestCaseParams.INPUT, LLMTestCaseParams.ACTUAL_OUTPUT],
        model=gemini_ai_model
    )

    test_case = LLMTestCase(
        input=prompt,
        actual_output=cleaned_output
    )

    assert_test(test_case, [valid_yaml_metric])
