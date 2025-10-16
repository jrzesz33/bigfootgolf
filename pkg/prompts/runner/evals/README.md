# Evaluation Modules

This directory contains all evaluation implementations for the golf agent runner.

## Creating a New Evaluator

### 1. Create Your Evaluator File

Create a new Python file in this directory (e.g., `my_evaluator.py`):

```python
"""
My Custom Evaluator

Brief description of what this evaluator does.
"""

from typing import List
from .base import BaseEvaluator, DatasetRecord, EvaluationResult


class MyCustomEvaluator(BaseEvaluator):
    """Evaluates responses for [your criterion]"""

    def __init__(self, config):
        super().__init__(name="MyCustom", config=config)
        # Initialize any models, clients, or resources

    def evaluate(self, records: List[DatasetRecord]) -> List[DatasetRecord]:
        """
        Main evaluation logic

        Args:
            records: List of DatasetRecord objects to evaluate

        Returns:
            Same list with evaluation results added
        """
        for record in records:
            # Your evaluation logic here
            # Example: call an API, use ML model, apply rules, etc.

            # Create result
            result = EvaluationResult(
                eval_name=self.name,
                label="your_label",  # e.g., "pass", "fail", "good", "bad"
                score=1.0,  # Numeric score (typically 0-1)
                explanation="Why this label was chosen",
                metadata={
                    "any": "additional",
                    "context": "you want to save"
                }
            )

            # Add result to record
            record.add_eval_result(result)

        return records
```

### 2. Optional: Implement Lifecycle Hooks

```python
class MyCustomEvaluator(BaseEvaluator):
    def pre_evaluate(self, records: List[DatasetRecord]) -> List[DatasetRecord]:
        """
        Called before evaluation - use for filtering or preprocessing

        Example: Filter out records that shouldn't be evaluated
        """
        valid_records = []
        for record in records:
            if self._should_evaluate(record):
                valid_records.append(record)
        return valid_records

    def evaluate(self, records: List[DatasetRecord]) -> List[DatasetRecord]:
        """Main evaluation logic (required)"""
        # Your implementation
        return records

    def post_evaluate(self, records: List[DatasetRecord]) -> List[DatasetRecord]:
        """
        Called after evaluation - use for cleanup or additional logging

        Example: Print detailed statistics
        """
        self.print_summary(records)
        return records
```

### 3. Register Your Evaluator

Add to `main_refactored.py`:

```python
from evals.my_evaluator import MyCustomEvaluator

AVAILABLE_EVALUATORS = {
    "hallucination": HallucinationEvaluator,
    "relevance": RelevanceEvaluator,
    "mycustom": MyCustomEvaluator,  # Add here
}
```

### 4. Run Your Evaluator

```bash
python main_refactored.py --evals mycustom
```

## Evaluator Examples

### Simple Rule-Based Evaluator

```python
class LengthEvaluator(BaseEvaluator):
    """Evaluates if responses are appropriate length"""

    def evaluate(self, records: List[DatasetRecord]) -> List[DatasetRecord]:
        for record in records:
            response_length = len(record.response)

            # Apply rules
            if response_length < 50:
                label = "too_short"
                score = 0.0
            elif response_length > 1000:
                label = "too_long"
                score = 0.5
            else:
                label = "appropriate"
                score = 1.0

            result = EvaluationResult(
                eval_name=self.name,
                label=label,
                score=score,
                explanation=f"Response length: {response_length} characters"
            )
            record.add_eval_result(result)

        return records
```

### API-Based Evaluator

```python
import requests

class SentimentEvaluator(BaseEvaluator):
    """Evaluates response sentiment using external API"""

    def __init__(self, config):
        super().__init__(name="Sentiment", config=config)
        self.api_endpoint = "https://api.sentimentapi.com/analyze"

    def evaluate(self, records: List[DatasetRecord]) -> List[DatasetRecord]:
        for record in records:
            # Call external API
            response = requests.post(
                self.api_endpoint,
                json={"text": record.response}
            )
            sentiment_data = response.json()

            result = EvaluationResult(
                eval_name=self.name,
                label=sentiment_data["sentiment"],  # "positive", "negative", "neutral"
                score=sentiment_data["confidence"],
                explanation=sentiment_data.get("explanation", ""),
                metadata={"api_version": sentiment_data.get("version")}
            )
            record.add_eval_result(result)

        return records
```

### LLM-Based Evaluator Template

```python
import boto3
from phoenix.evals import BedrockModel, llm_classify
import pandas as pd

class CustomLLMEvaluator(BaseEvaluator):
    """Template for LLM-based evaluators using Phoenix + Bedrock"""

    EVALUATION_TEMPLATE = """
Your evaluation prompt here.

[BEGIN DATA]
Question: {input}
Answer: {output}
Reference: {reference}  # Optional
[END DATA]

Your classification instructions here.

Respond with one of: label1, label2, label3
"""

    def __init__(self, config):
        super().__init__(name="CustomLLM", config=config)
        self.model = None

    def _initialize_model(self):
        if self.model is not None:
            return

        session = boto3.Session(region_name=self.config.aws.region)
        bedrock_client = session.client("bedrock-runtime")

        self.model = BedrockModel(
            model_id=self.config.aws.model_id,
            client=bedrock_client,
            temperature=0.0
        )

    def evaluate(self, records: List[DatasetRecord]) -> List[DatasetRecord]:
        self._initialize_model()

        # Prepare DataFrame
        eval_data = []
        for record in records:
            eval_data.append({
                "input": record.query,
                "output": record.response,
                "reference": record.reference
            })

        df = pd.DataFrame(eval_data)

        # Run Phoenix evaluation
        eval_results = llm_classify(
            dataframe=df,
            model=self.model,
            template=self.EVALUATION_TEMPLATE,
            rails=["label1", "label2", "label3"],
            provide_explanation=True
        )

        # Add results to records
        for i, (idx, result) in enumerate(eval_results.iterrows()):
            if i < len(records):
                eval_result = EvaluationResult(
                    eval_name=self.name,
                    label=result["label"],
                    score=float(result["score"]),
                    explanation=result.get("explanation", "")
                )
                records[i].add_eval_result(eval_result)

        return records
```

## Best Practices

### 1. Clear Naming
- Use descriptive evaluator names (e.g., `ToxicityEvaluator`, not `Eval1`)
- Use consistent label values (e.g., "pass"/"fail", "good"/"bad")

### 2. Comprehensive Results
- Always provide explanations
- Include relevant metadata
- Use meaningful score ranges

### 3. Error Handling
```python
def evaluate(self, records: List[DatasetRecord]) -> List[DatasetRecord]:
    for record in records:
        try:
            # Evaluation logic
            result = EvaluationResult(...)
            record.add_eval_result(result)
        except Exception as e:
            # Log error and add error result
            print(f"Error evaluating record: {e}")
            result = EvaluationResult(
                eval_name=self.name,
                label="error",
                score=0.0,
                explanation=f"Evaluation failed: {str(e)}"
            )
            record.add_eval_result(result)
    return records
```

### 4. Resource Management
```python
class MyEvaluator(BaseEvaluator):
    def __init__(self, config):
        super().__init__(name="MyEval", config=config)
        self.client = None  # Lazy initialization

    def _ensure_initialized(self):
        if self.client is None:
            self.client = create_expensive_resource()

    def evaluate(self, records):
        self._ensure_initialized()
        # Use self.client
        return records
```

### 5. Testability
```python
class MyEvaluator(BaseEvaluator):
    def _evaluate_single(self, record: DatasetRecord) -> EvaluationResult:
        """Evaluate a single record - easier to unit test"""
        # Logic for single record
        return EvaluationResult(...)

    def evaluate(self, records: List[DatasetRecord]) -> List[DatasetRecord]:
        for record in records:
            result = self._evaluate_single(record)
            record.add_eval_result(result)
        return records
```

## Available Evaluators

| Evaluator | Labels | Description |
|-----------|--------|-------------|
| `hallucination` | factual, hallucinated | Detects unsupported information |
| `relevance` | relevant, irrelevant | Checks if response addresses query |

## Common Patterns

### Filtering in pre_evaluate()
```python
def pre_evaluate(self, records: List[DatasetRecord]) -> List[DatasetRecord]:
    # Only evaluate records with responses
    return [r for r in records if r.response and not r.response.startswith("ERROR")]
```

### Batching for Efficiency
```python
def evaluate(self, records: List[DatasetRecord]) -> List[DatasetRecord]:
    # Batch API calls for efficiency
    BATCH_SIZE = 10
    for i in range(0, len(records), BATCH_SIZE):
        batch = records[i:i+BATCH_SIZE]
        results = self._evaluate_batch(batch)
        # Apply results...
    return records
```

### Progressive Evaluation
```python
def evaluate(self, records: List[DatasetRecord]) -> List[DatasetRecord]:
    # Show progress for long evaluations
    total = len(records)
    for i, record in enumerate(records):
        result = self._evaluate_single(record)
        record.add_eval_result(result)
        if (i + 1) % 10 == 0:
            print(f"  Progress: {i+1}/{total}")
    return records
```

## Testing Your Evaluator

Create a simple test:

```python
# test_my_evaluator.py
from evals.base import DatasetRecord
from evals.my_evaluator import MyCustomEvaluator
from config import EvalRunnerConfig

def test_my_evaluator():
    config = EvalRunnerConfig.from_env()

    records = [
        DatasetRecord(
            query="Test query",
            reference="Test reference",
            response="Test response"
        )
    ]

    evaluator = MyCustomEvaluator(config)
    results = evaluator.run(records)

    assert len(results) == 1
    assert len(results[0].eval_results) > 0
    print("Evaluator test passed!")

if __name__ == "__main__":
    test_my_evaluator()
```

Run test:
```bash
python test_my_evaluator.py
```
