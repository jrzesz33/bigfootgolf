"""
Hallucination evaluator using Phoenix and AWS Bedrock

Detects whether agent responses contain factually incorrect information
not supported by the reference text.
"""

import boto3
import pandas as pd
from typing import List
from phoenix.evals import BedrockModel, llm_classify

from .base import BaseEvaluator, DatasetRecord, EvaluationResult


class HallucinationEvaluator(BaseEvaluator):
    """
    Evaluates agent responses for hallucinations using Phoenix + Bedrock

    Uses Phoenix's hallucination template to compare agent responses
    against reference text and classify as "factual" or "hallucinated"
    """

    HALLUCINATION_TEMPLATE = """
You are comparing a reference text to a question and answer from an AI assistant.

[BEGIN REFERENCE TEXT]
{reference}
[END REFERENCE TEXT]

[BEGIN DATA]
************
[Question]: {input}
************
[AI Assistant Answer]: {output}
[END DATA]

Compare the AI assistant's answer to the reference text. Does the AI assistant's answer contain factually incorrect information or hallucinations that are not supported by the reference text?

Your response must be a single word, either "factual" or "hallucinated".

factual - The answer is supported by the reference text
hallucinated - The answer contains information not supported by the reference text
"""

    def __init__(self, config):
        """
        Initialize hallucination evaluator

        Args:
            config: EvalRunnerConfig with AWS and other settings
        """
        super().__init__(name="Hallucination", config=config)
        self.model = None

    def _initialize_model(self):
        """Initialize Bedrock model (lazy initialization)"""
        if self.model is not None:
            return

        print(f"Initializing Bedrock model: {self.config.aws.model_id}...")

        # Create boto3 session
        session_kwargs = {"region_name": self.config.aws.region}
        if self.config.aws.access_key_id and self.config.aws.secret_access_key:
            session_kwargs["aws_access_key_id"] = self.config.aws.access_key_id
            session_kwargs["aws_secret_access_key"] = self.config.aws.secret_access_key

        session = boto3.Session(**session_kwargs)
        bedrock_client = session.client("bedrock-runtime")

        # Initialize Phoenix Bedrock model
        self.model = BedrockModel(
            model_id=self.config.aws.model_id,
            client=bedrock_client,
            temperature=0.0
        )

        print(f"Model initialized successfully")

    def pre_evaluate(self, records: List[DatasetRecord]) -> List[DatasetRecord]:
        """Filter out records without responses"""
        valid_records = []
        for record in records:
            if record.response and not record.response.startswith("ERROR"):
                valid_records.append(record)
            else:
                print(f"  Skipping record with no valid response: {record.query[:50]}...")

        print(f"  Filtered to {len(valid_records)}/{len(records)} valid records")
        return valid_records

    def evaluate(self, records: List[DatasetRecord]) -> List[DatasetRecord]:
        """
        Evaluate records for hallucinations

        Args:
            records: List of DatasetRecord objects

        Returns:
            Same list with evaluation results added
        """
        if not records:
            print("  No records to evaluate")
            return records

        # Initialize model
        self._initialize_model()

        # Prepare DataFrame for Phoenix evaluation
        eval_data = []
        for record in records:
            eval_data.append({
                "input": record.query,
                "reference": record.reference,
                "output": record.response
            })

        df = pd.DataFrame(eval_data)

        # Run Phoenix evaluation
        print(f"  Evaluating {len(df)} responses for hallucinations...")

        eval_results = llm_classify(
            dataframe=df,
            model=self.model,
            template=self.HALLUCINATION_TEMPLATE,
            rails=["factual", "hallucinated"],
            provide_explanation=True
        )

        print("\n  Evaluation Results:")
        print(eval_results[["label", "score", "explanation"]])

        # Add results to records
        for i, (idx, result) in enumerate(eval_results.iterrows()):
            if i < len(records):
                eval_result = EvaluationResult(
                    eval_name=self.name,
                    label=result["label"],
                    score=float(result["score"]),
                    explanation=result.get("explanation", ""),
                    metadata={
                        "model_id": self.config.aws.model_id,
                        "template": "hallucination"
                    }
                )
                records[i].add_eval_result(eval_result)

        return records

    def post_evaluate(self, records: List[DatasetRecord]) -> List[DatasetRecord]:
        """Print summary after evaluation"""
        self.print_summary(records)
        return records
