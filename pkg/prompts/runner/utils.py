"""
Utility functions for the evaluation runner
"""

import json
import os
import time
import requests
from datetime import datetime
from typing import List, Dict, Any
from jwt import encode as jwt_encode

from config import EvalRunnerConfig
from evals.base import DatasetRecord


def generate_jwt_token(config: EvalRunnerConfig) -> str:
    """
    Generate JWT token for API authentication

    Args:
        config: EvalRunnerConfig with auth settings

    Returns:
        JWT token string
    """
    token_payload = {
        "user_id": config.auth.eval_user_id,
        "email": config.auth.eval_user_email,
        "elev": True,
        "exp": int(time.time()) + (config.auth.token_expiry_hours * 6600),
        "iat": int(time.time())
    }

    token = jwt_encode(token_payload, config.auth.jwt_secret, algorithm="HS256")
    return token


def load_dataset(config: EvalRunnerConfig, filename: str = "golf_agent_dataframes.json") -> List[Dict[str, Any]]:
    """
    Load dataset from JSON file

    Args:
        config: EvalRunnerConfig with path settings
        filename: Dataset filename

    Returns:
        List of dataset records as dictionaries
    """
    dataset_path = os.path.join(config.paths.datasets_dir, filename)

    with open(dataset_path, 'r') as f:
        dataset = json.load(f)

    return dataset


def save_dataset_with_responses(
    config: EvalRunnerConfig,
    dataset: List[Dict[str, Any]],
    run_timestamp: str
) -> str:
    """
    Save dataset with responses to dated JSON file

    Args:
        config: EvalRunnerConfig with path settings
        dataset: Dataset with responses
        run_timestamp: Timestamp string for the run

    Returns:
        Path to saved file
    """
    filename = f"responses_{run_timestamp}.json"
    output_path = os.path.join(config.paths.results_dir, filename)

    with open(output_path, 'w') as f:
        json.dump(dataset, f, indent=2)

    return output_path


def save_eval_results(
    config: EvalRunnerConfig,
    records: List[DatasetRecord],
    run_timestamp: str
) -> str:
    """
    Save evaluation results to dated JSON file

    Args:
        config: EvalRunnerConfig with path settings
        records: List of evaluated DatasetRecord objects
        run_timestamp: Timestamp string for the run

    Returns:
        Path to saved file
    """
    filename = f"eval_results_{run_timestamp}.json"
    output_path = os.path.join(config.paths.results_dir, filename)

    # Convert records to dict format
    results = [record.to_dict() for record in records]

    with open(output_path, 'w') as f:
        json.dump(results, f, indent=2)

    return output_path


def call_chat_api(
    config: EvalRunnerConfig,
    query: str,
    token: str
) -> str:
    """
    Call the chat API with a query

    Args:
        config: EvalRunnerConfig with API settings
        query: User query to send
        token: JWT authentication token

    Returns:
        Agent response string
    """
    session = requests.Session()
    headers = {
        "Authorization": f"Bearer {token}",
        "Content-Type": "application/json",
        "Accept": "application/json",
    }
    session.headers.update(headers)

    # Detailed logging for troubleshooting
    print("\n  === HTTP REQUEST DEBUG INFO ===")
    print(f"  Endpoint: {config.api.chat_endpoint}")
    print(f"  Method: POST")
    print(f"  Headers:")
    for key, value in headers.items():
        if key == "Authorization":
            # Show first and last 10 chars of token for security
            token_preview = f"{value[:17]}...{value[-10:]}" if len(value) > 27 else value
            print(f"    {key}: {token_preview}")
        else:
            print(f"    {key}: {value}")
    
    reqBody = {"message": query,
               "max_tokens": 4096,
               "temperature": 0.7,
               "conversation_history": [{
                   "tool_calls": None, 
                   "role": "user",
                   "content": query
               }]
                }
    print(f"  Request Body: {json.dumps(reqBody, indent=4)}")
    print(f"  Full Token (first 50 chars): {token[:50]}...")
    print(f"  Token Length: {len(token)} characters")
    print("  ================================\n")

    response = session.post(
        config.api.chat_endpoint,
        json=reqBody
    )

    # Log response details
    print(f"  Response Status Code: {response.status_code}")
    print(f"  Response Headers: {dict(response.headers)}")

    if response.status_code != 200:
        print(f"  Response Body: {response.text}")

    response.raise_for_status()
    response_data = response.json()

    agent_response = response_data.get("response", "") if isinstance(response_data, dict) else str(response_data)
    return agent_response


def get_run_timestamp() -> str:
    """
    Get formatted timestamp for the current run

    Returns:
        Timestamp string in format YYYY-MM-DD_HH-MM-SS
    """
    return datetime.now().strftime("%Y-%m-%d_%H-%M-%S")


def print_run_summary(records: List[DatasetRecord], run_timestamp: str):
    """
    Print summary of evaluation run

    Args:
        records: List of evaluated DatasetRecord objects
        run_timestamp: Timestamp of the run
    """
    print("\n" + "=" * 80)
    print("EVALUATION RUN SUMMARY")
    print("=" * 80)
    print(f"Run Timestamp: {run_timestamp}")
    print(f"Total Records: {len(records)}")

    # Collect all unique eval names
    eval_names = set()
    for record in records:
        for result in record.eval_results:
            eval_names.add(result.eval_name)

    print(f"Evaluations Run: {', '.join(sorted(eval_names))}")

    # Print per-eval summaries
    for eval_name in sorted(eval_names):
        eval_results = []
        for record in records:
            for result in record.eval_results:
                if result.eval_name == eval_name:
                    eval_results.append(result)

        if eval_results:
            avg_score = sum(r.score for r in eval_results) / len(eval_results)
            label_counts = {}
            for result in eval_results:
                label_counts[result.label] = label_counts.get(result.label, 0) + 1

            print(f"\n  {eval_name}:")
            print(f"    Average Score: {avg_score:.2f}")
            print(f"    Label Distribution:")
            for label, count in sorted(label_counts.items()):
                percentage = (count / len(eval_results) * 100)
                print(f"      - {label}: {count} ({percentage:.1f}%)")

    print("\n" + "=" * 80)
