"""
Hallucination evaluator using LiteLLM with AWS Bedrock

Detects whether agent responses contain factually incorrect information
not supported by the reference text.
"""
import os
import boto3
from typing import List
import pandas as pd
from phoenix.client import Client
from phoenix.evals.utils import to_annotation_dataframe
from phoenix.evals.metrics.hallucination import HallucinationEvaluator
from phoenix.evals.llm import LLM
# from phoenix.evals import (
#     BedrockModel,
#     HALLUCINATION_PROMPT_RAILS_MAP,
#     HALLUCINATION_PROMPT_TEMPLATE,
#     download_benchmark_dataset,
#     llm_classify,
# )
from .base import BaseEvaluator, DatasetRecord, EvaluationResult


class HallucinationEvaluator(BaseEvaluator):
    """
    Evaluates agent responses for hallucinations using LiteLLM + Bedrock

    Uses a custom hallucination template to compare agent responses
    against reference text and classify as "factual" or "hallucinated"
    """

    def __init__(self, config):
        """
        Initialize hallucination evaluator

        Args:
            config: EvalRunnerConfig with AWS and other settings
        """
        super().__init__(name="Hallucination", config=config)
        self._setup_aws_env()

    def _setup_aws_env(self):
        """Setup AWS environment variables for LiteLLM"""
        self.config.aws.region = "us-east-1"
        self.config.aws.aws_session = os.getenv("AWS_BEARER_TOKEN_BEDROCK")
        self.config.aws.model_id = "us.amazon.nova-pro-v1:0"
        print(f"  Model: {self.config.aws}")
        #create the session
        session = boto3.Session(
            aws_session_token=self.config.aws.aws_session,
            region_name=self.config.aws.region
        )
        # create the model
        bedrock_client = session.client("bedrock-runtime")
        # self.eval_model = BedrockModel(
        #     model_id="us.amazon.nova-pro-v1:0",
        #     client=bedrock_client,
        #     temperature=0.0
        # )
        self.llm = LLM(
            provider="openai",
            model="us.amazon.nova-pro-v1:0",
            #api_key=self.config.aws.aws_session,
            client="openai",
            base_url="http://localhost:3002",
            #aws_session_token=self.config.aws.aws_session,
        )

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

        print(f"  Evaluating {len(records)} responses for hallucinations...")

        for i, record in enumerate(records):
            print(f"\n  Evaluating {i+1}/{len(records)}: {record.query[:50]}...")

            # Format prompt with record data
            eval_input = {
                "input": record.query,
                "context": record.reference,
                "output": record.response,
            }

            # df = pd.DataFrame(eval_input)
            # rails = list(HALLUCINATION_PROMPT_RAILS_MAP.values())
            # scores = llm_classify(
            #     dataframe=df, 
            #     template=HALLUCINATION_PROMPT_TEMPLATE, 
            #     model=self.eval_model, 
            #     rails=rails,
            #     provide_explanation=True, #optional to generate explanations for the value produced by the eval LLM
            # )

            hallucination_eval = HallucinationEvaluator(self.llm)
            scores = hallucination_eval.evaluate(eval_input)

            hallucination_annotations = to_annotation_dataframe(scores)
            
            client = Client(base_url="http://localhost:6006")
            client.spans.log_span_annotations_dataframe(dataframe=hallucination_annotations)

        return records

    def post_evaluate(self, records: List[DatasetRecord]) -> List[DatasetRecord]:
        """Print summary after evaluation"""
        self.print_summary(records)
        return records
