"""
Halluc TEst
"""

import os
import boto3
import pandas as pd
from phoenix.evals import (
    BedrockModel,
    HALLUCINATION_PROMPT_RAILS_MAP,
    HALLUCINATION_PROMPT_TEMPLATE,
    download_benchmark_dataset,
    llm_classify,
)
from phoenix.evals.metrics.hallucination import HallucinationEvaluator
from phoenix.evals.llm import LLM
#from phoenix.evals.llm import show_provider_availability, LLM
#from phoenix.evals.metrics import HallucinationEvaluator


bt = os.getenv("AWS_BEARER_TOKEN_BEDROCK")
#session_kwargs = {"region_name": "us-east-1"}
#session_kwargs["aws_access_key_id"] = self.config.aws.access_key_id
#session_kwargs["aws_secret_access_key"] = self.config.aws.secret_access_key
#session_kwargs["aws_bearer_token_bedrock"] = bt
#session_kwargs["api_key"] = bt

#session = boto3.Session(**session_kwargs)
session = boto3.Session(
    aws_session_token=bt,
    region_name="us-east-1"
)

bedrock_client = session.client("bedrock-runtime")
eval_model = BedrockModel(
    model_id="us.amazon.nova-pro-v1:0",
    session=session,
    #client=bedrock_client,
    temperature=0.0
)
os.environ['AWS_BEARER_TOKEN_BEDROCK'] = bt
print(bt)
llm_judge = LLM(
        provider="bedrock",
        model="us.amazon.nova-pro-v1:0",
        api_key=bt,
        client="litellm",
        session=session
        #base_url="http://localhost:3002",
        #aws_session_token=self.config.aws.aws_session,
    )

#llm_judge = LLM(model="amazon.nova-pro-v1:0", provider="bedrock")

#show_provider_availability()

#bt = os.getenv("AWS_BEARER_TOKEN_BEDROCK")
#print("Bedrock Token: ", bt)
#os.environ["AWS_BEARER_TOKEN_BEDROCK"] = bt

# initialize LLM and evaluator 
#llm = LLM(
#    provider="litellm",
#    model="bedrock/us.amazon.nova-pro-v1:0",
#)

hallucination = HallucinationEvaluator(llm_judge)

# use the .describe() method to inspect the input_schema of any evaluator
# print(hallucination.describe())

# let's test on one example
""" eval_input = [{
    "input": "Where is the Eiffel Tower located?",
    "reference": "The Eiffel Tower is located in Paris, France. It was constructed in 1889 as the entrance arch to the 1889 World's Fair.",
    "output": "The Eiffel Tower is located in Paris, France.",
}] """
eval_input = {
    "input": "Where is the Eiffel Tower located?",
    "context": "The Eiffel Tower is located in Paris, France. It was constructed in 1889 as the entrance arch to the 1889 World's Fair.",
    "output": "The Eiffel Tower is located in Paris, France.",
}
# print(HALLUCINATION_PROMPT_TEMPLATE)
#df = pd.DataFrame(eval_input)
# rails = list(HALLUCINATION_PROMPT_RAILS_MAP.values())
# scores = llm_classify(
#     dataframe=df, 
#     template=HALLUCINATION_PROMPT_TEMPLATE, 
#     model=eval_model, 
#     rails=rails,
#     provide_explanation=True, #optional to generate explanations for the value produced by the eval LLM
# )


scores = hallucination.evaluate(eval_input)
print(scores)
