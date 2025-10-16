from phoenix.client import Client
from phoenix.client.types import PromptVersion
import os


prompt_name = "golf_agent_system"

client = Client(
 # endpoint="https://my-phoenix.com",
)

def load_system_prompt():
    prompt_path = os.path.join(os.path.dirname(__file__), prompt_name + ".txt")
    print(prompt_path)
    with open(prompt_path, "r") as f:
        return f.read()

prompt_template = load_system_prompt()
prompt_version= PromptVersion([
            {
                "role": "system", 
                "content": prompt_template
            },
            {
                "role": "user", 
                "content": "{{question}}"
            }
        ],
        model_name="anthropic.claude-sonnet-4-5-20250929-v1:0",
        model_provider="AWS")

prompt = client.prompts.create(
    name=prompt_name,
    version=prompt_version,
)
print(f"Prompt '{prompt_name}' saved successfully.")
