"""
Evaluation modules for the golf agent evaluation runner

This package contains all evaluation implementations that can be run
against agent responses.
"""

from .base import BaseEvaluator, EvaluationResult
from .hallucination import HallucinationEvaluator

__all__ = [
    "BaseEvaluator",
    "EvaluationResult",
    "HallucinationEvaluator"
]
