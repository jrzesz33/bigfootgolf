# Golf Agent Evaluation Runner (Refactored)

Modular, extensible evaluation system for testing golf booking agent responses with support for multiple evaluators and dated output files.

## 🏗️ Architecture

```
runner/
├── main_refactored.py       # Main entry point (refactored)
├── config.py                 # Configuration management
├── utils.py                  # Utility functions
├── evals/                    # Evaluation modules
│   ├── __init__.py
│   ├── base.py              # Base evaluator class
│   ├── hallucination.py     # Hallucination evaluator
│   └── relevance.py         # Relevance evaluator (example)
└── results/                  # Auto-generated dated results
    ├── responses_2025-10-16_14-30-00.json
    └── eval_results_2025-10-16_14-30-00.json
```

## 🚀 Features

- **Modular Design**: Easy to add new evaluators
- **Dated Output Files**: Results saved with timestamps
- **Multiple Evaluators**: Run any combination of evaluators
- **Extensible Base Class**: Simple inheritance model for custom evaluators
- **CLI Interface**: Command-line arguments for flexibility
- **Configuration Management**: Centralized config with validation

## 📦 Installation

```bash
cd pkg/prompts/runner
pip install -r requirements.txt
```

## ⚙️ Configuration

Set environment variables (see `.env.example`):

```bash
# Required
export DB_ADMIN="your-neo4j-password"
export JWT_SECRET="your-jwt-secret"
export AWS_REGION="us-east-1"

# Optional (with defaults)
export DB_URI="bolt://localhost:7687"
export DB_USER="neo4j"
export API_BASE_URL="http://localhost:8000"
export BEDROCK_MODEL_ID="anthropic.claude-sonnet-4-5-20250929-v1:0"
```

## 🎯 Usage

### Run All Evaluators

```bash
python main_refactored.py
```

### Run Specific Evaluators

```bash
python main_refactored.py --evals hallucination,relevance
```

### Use Custom Dataset

```bash
python main_refactored.py --dataset my_custom_dataset.json
```

### List Available Evaluators

```bash
python main_refactored.py --list-evals
```

## 📊 Output Files

Results are automatically saved with timestamps in the `results/` directory:

### Response Files
Format: `responses_YYYY-MM-DD_HH-MM-SS.json`

```json
[
  {
    "query": "What is the weather?",
    "reference": "Expected behavior...",
    "response": "Agent response..."
  }
]
```

### Evaluation Results
Format: `eval_results_YYYY-MM-DD_HH-MM-SS.json`

```json
[
  {
    "query": "What is the weather?",
    "reference": "Expected behavior...",
    "response": "Agent response...",
    "evaluations": [
      {
        "eval_name": "Hallucination",
        "label": "factual",
        "score": 1.0,
        "explanation": "Response is supported by reference",
        "metadata": {
          "model_id": "anthropic.claude-sonnet-4-5-20250929-v1:0",
          "template": "hallucination"
        },
        "timestamp": "2025-10-16T14:30:00.123456"
      }
    ]
  }
]
```

## 🔧 Adding New Evaluators

### 1. Create Evaluator Class

Create a new file in `evals/` directory (e.g., `evals/toxicity.py`):

```python
from typing import List
from .base import BaseEvaluator, DatasetRecord, EvaluationResult

class ToxicityEvaluator(BaseEvaluator):
    """Evaluates responses for toxic or harmful content"""

    def __init__(self, config):
        super().__init__(name="Toxicity", config=config)
        # Initialize any models or resources

    def evaluate(self, records: List[DatasetRecord]) -> List[DatasetRecord]:
        """Implement your evaluation logic"""
        for record in records:
            # Your evaluation logic here
            result = EvaluationResult(
                eval_name=self.name,
                label="safe",  # or "toxic"
                score=1.0,
                explanation="Response contains no toxic content"
            )
            record.add_eval_result(result)
        return records
```

### 2. Register Evaluator

Add to `main_refactored.py`:

```python
from evals.toxicity import ToxicityEvaluator

AVAILABLE_EVALUATORS = {
    "hallucination": HallucinationEvaluator,
    "relevance": RelevanceEvaluator,
    "toxicity": ToxicityEvaluator,  # Add your evaluator
}
```

### 3. Run Your Evaluator

```bash
python main_refactored.py --evals toxicity
```

## 📝 Evaluator Lifecycle Hooks

The `BaseEvaluator` provides lifecycle hooks:

```python
class MyEvaluator(BaseEvaluator):
    def pre_evaluate(self, records):
        """Called before evaluation - use for filtering/preprocessing"""
        return records

    def evaluate(self, records):
        """Main evaluation logic - REQUIRED"""
        return records

    def post_evaluate(self, records):
        """Called after evaluation - use for cleanup/logging"""
        return records
```

## 🔍 Available Evaluators

### Hallucination Evaluator
```bash
python main_refactored.py --evals hallucination
```

Detects factually incorrect information not supported by reference text.
- Labels: `factual`, `hallucinated`
- Uses: Phoenix + AWS Bedrock
- Template: Compares response against reference

### Relevance Evaluator
```bash
python main_refactored.py --evals relevance
```

Evaluates whether responses appropriately address the query.
- Labels: `relevant`, `irrelevant`
- Uses: Phoenix + AWS Bedrock
- Template: Checks if answer addresses question

## 📈 Example Output

```
================================================================================
GOLF AGENT EVALUATION RUNNER
================================================================================
Run Timestamp: 2025-10-16_14-30-00

[1/5] Loading dataset...
  Loaded 3 test cases from golf_agent_dataframes.json

[2/5] Generating JWT token...
  Generated token for user: eval-runner

[3/5] Calling chat API for queries...
  Query 1/3: What is the weather at the golf course over the weekend?
  Response: Based on the weather forecast...

[4/5] Preparing records for evaluation...
  Prepared 3 records

[5/5] Running evaluators...
  Evaluators to run: hallucination, relevance

================================================================================
Running Hallucination Evaluator
================================================================================
  Filtered to 3/3 valid records
  Evaluating 3 responses for hallucinations...

  Hallucination Summary:
    Total evaluations: 3
    Average score: 1.00
    Label distribution:
      - factual: 3 (100.0%)

================================================================================
Running Relevance Evaluator
================================================================================
  Filtered to 3/3 valid records
  Evaluating 3 responses for relevance...

  Relevance Summary:
    Total evaluations: 3
    Average score: 1.00
    Label distribution:
      - relevant: 3 (100.0%)

  Saved evaluation results to: results/eval_results_2025-10-16_14-30-00.json

================================================================================
EVALUATION RUN SUMMARY
================================================================================
Run Timestamp: 2025-10-16_14-30-00
Total Records: 3
Evaluations Run: Hallucination, Relevance

  Hallucination:
    Average Score: 1.00
    Label Distribution:
      - factual: 3 (100.0%)

  Relevance:
    Average Score: 1.00
    Label Distribution:
      - relevant: 3 (100.0%)

================================================================================
EVALUATION COMPLETE
================================================================================
```

## 🧪 Testing Multiple Evaluators

Run all evaluators in sequence:

```bash
python main_refactored.py --evals hallucination,relevance
```

Each evaluator:
1. Receives the same `DatasetRecord` objects
2. Adds its own `EvaluationResult` to each record
3. Returns the records for the next evaluator
4. Results accumulate in the `evaluations` array

## 🗂️ File Organization

### Old Structure (main.py)
- Monolithic single file
- Hard to extend
- No dated outputs

### New Structure (main_refactored.py)
- Modular evaluator system
- Easy to add new evaluators
- Automatic dated outputs
- Centralized configuration
- Reusable utilities

## 🔄 Migration Guide

To migrate from old `main.py` to new structure:

1. Update imports:
   ```bash
   # Old
   python main.py

   # New
   python main_refactored.py
   ```

2. Results now in `results/` directory with timestamps

3. Add custom evaluators in `evals/` directory

4. Use CLI args to control which evaluators run

## 🛠️ Troubleshooting

### Import Errors

Ensure you're in the runner directory:
```bash
cd /workspaces/golf_app/pkg/prompts/runner
python main_refactored.py
```

### Configuration Errors

Check environment variables:
```bash
python main_refactored.py  # Will show validation errors
```

### Evaluator Not Found

List available evaluators:
```bash
python main_refactored.py --list-evals
```

## 📚 API Reference

### DatasetRecord

```python
@dataclass
class DatasetRecord:
    query: str
    reference: str
    response: str
    eval_results: List[EvaluationResult]

    def add_eval_result(self, result: EvaluationResult)
    def to_dict(self) -> Dict[str, Any]
```

### EvaluationResult

```python
@dataclass
class EvaluationResult:
    eval_name: str
    label: str
    score: float
    explanation: Optional[str]
    metadata: Dict[str, Any]
    timestamp: str

    def to_dict(self) -> Dict[str, Any]
```

### BaseEvaluator

```python
class BaseEvaluator(ABC):
    def __init__(self, name: str, config: Any)
    def evaluate(self, records: List[DatasetRecord]) -> List[DatasetRecord]
    def pre_evaluate(self, records: List[DatasetRecord]) -> List[DatasetRecord]
    def post_evaluate(self, records: List[DatasetRecord]) -> List[DatasetRecord]
    def run(self, records: List[DatasetRecord]) -> List[DatasetRecord]
    def print_summary(self, records: List[DatasetRecord])
```

## 🎓 Best Practices

1. **One Evaluator Per Concern**: Create separate evaluators for different aspects (hallucination, relevance, toxicity, etc.)

2. **Use Pre/Post Hooks**: Implement `pre_evaluate()` for filtering, `post_evaluate()` for logging

3. **Descriptive Names**: Use clear eval names and label values

4. **Rich Explanations**: Provide detailed explanations in results

5. **Metadata**: Store additional context in the metadata field

6. **Error Handling**: Handle errors gracefully in evaluators

## 🤝 Contributing

To add a new evaluator:

1. Create file in `evals/` directory
2. Inherit from `BaseEvaluator`
3. Implement `evaluate()` method
4. Register in `AVAILABLE_EVALUATORS`
5. Update this README with documentation

## 📄 License

Part of the Golf Booking Application
