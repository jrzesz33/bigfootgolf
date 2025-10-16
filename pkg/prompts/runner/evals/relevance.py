"""
Relevance evaluator using Phoenix and AWS Bedrock

Evaluates whether agent responses are relevant and appropriate to the user query.
"""

import boto3
import pandas as pd
from typing import List
from phoenix.evals import BedrockModel, llm_classify

from .base import BaseEvaluator, DatasetRecord, EvaluationResult


class RelevanceEvaluator(BaseEvaluator):
    """
    Evaluates agent responses for relevance to the query

    Determines if the response appropriately addresses the user's question
    """

    RELEVANCE_TEMPLATE = """
You are evaluating how relevant an AI assistant's response is to a user's question.

[BEGIN DATA]
************
[Question]: {input}
************
[AI Assistant Answer]: {output}
[END DATA]

Does the AI assistant's answer appropriately address the user's question?
Consider:
- Does it answer what was asked?
- Is it on-topic?
- Does it provide useful information related to the question?

Your response must be a single word, either "relevant" or "irrelevant".

relevant - The answer appropriately addresses the question
irrelevant - The answer does not address the question or is off-topic
"""

    def __init__(self, config):
        """
        Initialize relevance evaluator

        Args:
            config: EvalRunnerConfig with AWS and other settings
        """
        super().__init__(name="Relevance", config=config)
        self.model = None

    def _initialize_model(self):
        """Initialize Bedrock model (lazy initialization)"""
        if self.model is not None:
            return

        print(f"Initializing Bedrock model: {self.config.aws.model_id}...")

        session_kwargs = {"region_name": self.config.aws.region}
        if self.config.aws.access_key_id and self.config.aws.secret_access_key:
            session_kwargs["aws_access_key_id"] = self.config.aws.access_key_id
            session_kwargs["aws_secret_access_key"] = self.config.aws.secret_access_key

        session = boto3.Session(**session_kwargs)
        bedrock_client = session.client("bedrock-runtime")

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

        print(f"  Filtered to {len(valid_records)}/{len(records)} valid records")
        return valid_records

    def evaluate(self, records: List[DatasetRecord]) -> List[DatasetRecord]:
        """
        Evaluate records for relevance

        Args:
            records: List of DatasetRecord objects

        Returns:
            Same list with evaluation results added
        """
        if not records:
            print("  No records to evaluate")
            return records

        self._initialize_model()

        eval_data = []
        for record in records:
            eval_data.append({
                "input": record.query,
                "output": record.response
            })

        df = pd.DataFrame(eval_data)

        print(f"  Evaluating {len(df)} responses for relevance...")

        eval_results = llm_classify(
            dataframe=df,
            model=self.model,
            template=self.RELEVANCE_TEMPLATE,
            rails=["relevant", "irrelevant"],
            provide_explanation=True
        )

        print("\n  Relevance Results:")
        print(eval_results[["label", "score", "explanation"]])

        for i, (idx, result) in enumerate(eval_results.iterrows()):
            if i < len(records):
                eval_result = EvaluationResult(
                    eval_name=self.name,
                    label=result["label"],
                    score=float(result["score"]),
                    explanation=result.get("explanation", ""),
                    metadata={
                        "model_id": self.config.aws.model_id,
                        "template": "relevance"
                    }
                )
                records[i].add_eval_result(eval_result)

        return records

    def post_evaluate(self, records: List[DatasetRecord]) -> List[DatasetRecord]:
        """Print summary after evaluation"""
        self.print_summary(records)
        return records
