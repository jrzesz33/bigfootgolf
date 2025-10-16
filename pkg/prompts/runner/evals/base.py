"""
Base evaluator class and evaluation result structures

All custom evaluators should inherit from BaseEvaluator
"""

from abc import ABC, abstractmethod
from dataclasses import dataclass, field
from typing import Any, Dict, List, Optional
from datetime import datetime


@dataclass
class EvaluationResult:
    """Result of a single evaluation"""
    eval_name: str
    label: str
    score: float
    explanation: Optional[str] = None
    metadata: Dict[str, Any] = field(default_factory=dict)
    timestamp: str = field(default_factory=lambda: datetime.now().isoformat())

    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary for JSON serialization"""
        return {
            "eval_name": self.eval_name,
            "label": self.label,
            "score": self.score,
            "explanation": self.explanation,
            "metadata": self.metadata,
            "timestamp": self.timestamp
        }


@dataclass
class DatasetRecord:
    """Represents a single test case in the dataset"""
    query: str
    reference: str
    response: str
    eval_results: List[EvaluationResult] = field(default_factory=list)

    def add_eval_result(self, result: EvaluationResult):
        """Add an evaluation result to this record"""
        self.eval_results.append(result)

    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary for JSON serialization"""
        return {
            "query": self.query,
            "reference": self.reference,
            "response": self.response,
            "evaluations": [r.to_dict() for r in self.eval_results]
        }


class BaseEvaluator(ABC):
    """
    Base class for all evaluators

    Custom evaluators should inherit from this class and implement
    the evaluate() method.
    """

    def __init__(self, name: str, config: Any = None):
        """
        Initialize evaluator

        Args:
            name: Name of this evaluator
            config: Configuration object (optional)
        """
        self.name = name
        self.config = config

    @abstractmethod
    def evaluate(self, records: List[DatasetRecord]) -> List[DatasetRecord]:
        """
        Evaluate a list of dataset records

        Args:
            records: List of DatasetRecord objects to evaluate

        Returns:
            List of DatasetRecord objects with evaluation results added
        """
        pass

    def pre_evaluate(self, records: List[DatasetRecord]) -> List[DatasetRecord]:
        """
        Hook called before evaluation (optional override)

        Can be used for filtering, preprocessing, etc.

        Args:
            records: List of DatasetRecord objects

        Returns:
            Filtered/preprocessed list of records
        """
        return records

    def post_evaluate(self, records: List[DatasetRecord]) -> List[DatasetRecord]:
        """
        Hook called after evaluation (optional override)

        Can be used for post-processing, logging, etc.

        Args:
            records: List of DatasetRecord objects with results

        Returns:
            Post-processed list of records
        """
        return records

    def run(self, records: List[DatasetRecord]) -> List[DatasetRecord]:
        """
        Run the complete evaluation pipeline

        Args:
            records: List of DatasetRecord objects

        Returns:
            List of DatasetRecord objects with evaluation results
        """
        print(f"\n{'='*80}")
        print(f"Running {self.name} Evaluator")
        print(f"{'='*80}")

        # Pre-evaluation hook
        records = self.pre_evaluate(records)

        # Run evaluation
        records = self.evaluate(records)

        # Post-evaluation hook
        records = self.post_evaluate(records)

        print(f"\nCompleted {self.name} evaluation on {len(records)} records\n")
        return records

    def print_summary(self, records: List[DatasetRecord]):
        """
        Print evaluation summary statistics

        Args:
            records: List of evaluated DatasetRecord objects
        """
        # Find results from this evaluator
        eval_results = []
        for record in records:
            for result in record.eval_results:
                if result.eval_name == self.name:
                    eval_results.append(result)

        if not eval_results:
            print(f"No {self.name} results found")
            return

        # Calculate statistics
        total = len(eval_results)
        avg_score = sum(r.score for r in eval_results) / total if total > 0 else 0

        # Count by label
        label_counts = {}
        for result in eval_results:
            label_counts[result.label] = label_counts.get(result.label, 0) + 1

        print(f"\n{self.name} Summary:")
        print(f"  Total evaluations: {total}")
        print(f"  Average score: {avg_score:.2f}")
        print(f"  Label distribution:")
        for label, count in sorted(label_counts.items()):
            percentage = (count / total * 100) if total > 0 else 0
            print(f"    - {label}: {count} ({percentage:.1f}%)")
