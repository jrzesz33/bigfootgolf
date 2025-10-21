#!/usr/bin/env python3
"""
Golf Agent Evaluation Runner

Refactored modular version with support for multiple evaluators and dated output files.

Usage:
    python main_refactored.py [--evals hallucination,relevance] [--dataset golf_agent_dataframes.json]
"""

import argparse
import sys
import time
from typing import List

from config import EvalRunnerConfig
from evals.base import DatasetRecord
from evals.hallucination import HallucinationEvaluator
from evals.relevance import RelevanceEvaluator
from utils import (
    generate_jwt_token,
    load_dataset,
    save_dataset_with_responses,
    save_eval_results,
    call_chat_api,
    get_run_timestamp,
    print_run_summary
)


# Registry of available evaluators
AVAILABLE_EVALUATORS = {
    "hallucination": HallucinationEvaluator,
    "relevance": RelevanceEvaluator
}


def run_evaluation_pipeline(
    config: EvalRunnerConfig,
    dataset_filename: str = "golf_agent_dataframes.json",
    evaluator_names: List[str] = None
):
    """
    Run the complete evaluation pipeline

    Args:
        config: Configuration object
        dataset_filename: Name of dataset file to load
        evaluator_names: List of evaluator names to run (default: all)
    """
    print("\n" + "=" * 80)
    print("GOLF AGENT EVALUATION RUNNER")
    print("=" * 80)

    # Get run timestamp for file naming
    run_timestamp = get_run_timestamp()
    print(f"Run Timestamp: {run_timestamp}\n")

    # Step 1: Load dataset
    print("[1/5] Loading dataset...")
    try:
        dataset = load_dataset(config, dataset_filename)
        print(f"  Loaded {len(dataset)} test cases from {dataset_filename}")

        # Check if dataset already has responses
        has_responses = all(record.get("response") for record in dataset)

        if has_responses:
            print(f"  Dataset already contains responses - skipping API calls")
            skip_api_calls = True
        else:
            print(f"  Dataset does not have responses - will call API")
            skip_api_calls = False

    except FileNotFoundError:
        print(f"  ERROR: Dataset not found: {dataset_filename}")
        return
    except Exception as e:
        print(f"  ERROR loading dataset: {e}")
        return

    if not skip_api_calls:
        # Step 2: Generate JWT token
        print("\n[2/5] Generating JWT token...")
        try:
            token = generate_jwt_token(config)
            print(f"  Generated token for user: {config.auth.eval_user_id}")
            print(f"  Token: {token}")
            print(f"  Secret: {config.auth.jwt_secret}")

        except Exception as e:
            print(f"  ERROR generating token: {e}")
            return

        # Step 3: Call chat API for each query
        print("\n[3/5] Calling chat API for queries...")
        for i, record in enumerate(dataset):
            query = record.get("query", "")
            print(f"\n  Query {i+1}/{len(dataset)}: {query}")

            try:
                response = call_chat_api(config, query, token)
                record["response"] = response
                print(f"  Response: {response}{'...' if len(response) > 100 else ''}")
            except Exception as e:
                print(f"  ERROR calling chat API: {e}")
                record["response"] = f"ERROR: {str(e)}"

            # Wait 5 seconds between requests to avoid rate limiting
            if i < len(dataset) - 1:  # Don't wait after the last request
                print(f"  Waiting 10 seconds before next request...")
                time.sleep(10)

        # Save dataset with responses
        responses_path = save_dataset_with_responses(config, dataset, run_timestamp)
        print(f"\n  Saved responses to: {responses_path}")
    else:
        print("\n[2/5] Skipping JWT token generation (responses already present)")
        print("\n[3/5] Skipping chat API calls (responses already present)")

    # Step 4: Convert to DatasetRecord objects
    print("\n[4/5] Preparing records for evaluation...")
    records = []
    for item in dataset:
        record = DatasetRecord(
            query=item.get("query", ""),
            reference=item.get("reference", ""),
            response=item.get("response", "")
        )
        records.append(record)
    print(f"  Prepared {len(records)} records")

    # Step 5: Run evaluators
    print("\n[5/5] Running evaluators...")

    # Determine which evaluators to run
    if evaluator_names is None:
        evaluator_names = list(AVAILABLE_EVALUATORS.keys())

    print(f"  Evaluators to run: {', '.join(evaluator_names)}")

    for eval_name in evaluator_names:
        if eval_name not in AVAILABLE_EVALUATORS:
            print(f"\n  WARNING: Unknown evaluator '{eval_name}', skipping...")
            continue

        try:
            # Instantiate evaluator
            evaluator_class = AVAILABLE_EVALUATORS[eval_name]
            evaluator = evaluator_class(config)

            # Run evaluation
            records = evaluator.run(records)

        except Exception as e:
            print(f"\n  ERROR running {eval_name} evaluator: {e}")
            import traceback
            traceback.print_exc()

    # Save final results
    results_path = save_eval_results(config, records, run_timestamp)
    print(f"\n  Saved evaluation results to: {results_path}")

    # Print summary
    print_run_summary(records, run_timestamp)

    print("\n" + "=" * 80)
    print("EVALUATION COMPLETE")
    print("=" * 80 + "\n")


def main():
    """Main entry point with CLI argument parsing"""
    parser = argparse.ArgumentParser(
        description="Golf Agent Evaluation Runner"
    )
    parser.add_argument(
        "--evals",
        type=str,
        default=None,
        help=f"Comma-separated list of evaluators to run (available: {', '.join(AVAILABLE_EVALUATORS.keys())})"
    )
    parser.add_argument(
        "--dataset",
        type=str,
        default="golf_agent_dataframes.json",
        help="Dataset filename to use (default: golf_agent_dataframes.json)"
    )
    parser.add_argument(
        "--list-evals",
        action="store_true",
        help="List available evaluators and exit"
    )

    args = parser.parse_args()

    # Handle --list-evals
    if args.list_evals:
        print("\nAvailable Evaluators:")
        for name, evaluator_class in AVAILABLE_EVALUATORS.items():
            print(f"  - {name}: {evaluator_class.__doc__.strip() if evaluator_class.__doc__ else 'No description'}")
        print()
        return

    # Load configuration
    config = EvalRunnerConfig.from_env()

    # Validate configuration
    errors = config.validate()
    if errors:
        print("Configuration errors:")
        for error in errors:
            print(f"  - {error}")
        print("\nPlease set required environment variables and try again.")
        sys.exit(1)

    # Parse evaluator names
    evaluator_names = None
    if args.evals:
        evaluator_names = [e.strip() for e in args.evals.split(",")]

    # Run evaluation pipeline
    try:
        run_evaluation_pipeline(
            config=config,
            dataset_filename=args.dataset,
            evaluator_names=evaluator_names
        )
    except KeyboardInterrupt:
        print("\n\nEvaluation interrupted by user")
        sys.exit(1)
    except Exception as e:
        print(f"\n\nFATAL ERROR: {e}")
        import traceback
        traceback.print_exc()
        sys.exit(1)


if __name__ == "__main__":
    main()
