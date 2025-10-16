from phoenix.evals.llm import show_provider_availability
from phoenix.client import Client
from pprint import pprint
import sys
import json
import os

# Initialize a phoenix client with your phoenix endpoint
# By default it will read from your environment variables
client = Client(
 # endpoint="https://my-phoenix.com",
)

# Pulling a prompt by name
prompt_name = "golf_agent_system"
version = sys.argv[1] if len(sys.argv) > 1 else None

if version:
    prompt = client.prompts.get(prompt_identifier=prompt_name, version=int(version))
else:
    prompt = client.prompts.get(prompt_identifier=prompt_name)


# Save prompt as JSON in versions folder
versions_dir = os.path.join(os.path.dirname(__file__), "versions")
os.makedirs(versions_dir, exist_ok=True)

version_id = prompt.version_id if hasattr(prompt, 'version_id') else (prompt.id if hasattr(prompt, 'id') else 'unknown')
output_path = os.path.join(versions_dir, f"{version_id}.json")

with open(output_path, "w") as f:
    json.dump(vars(prompt), f, indent=2, default=str)


show_provider_availability()

print(f"\nPrompt saved to: {output_path}")