"""
Configuration module for the evaluation runner

Centralizes all configuration settings and environment variables
"""

import os
from dataclasses import dataclass
from typing import Optional


@dataclass
class DatabaseConfig:
    """Neo4j database configuration"""
    uri: str
    user: str
    password: str

    @classmethod
    def from_env(cls):
        return cls(
            uri=os.getenv("DB_URI", "bolt://localhost:7687"),
            user=os.getenv("DB_USER", "neo4j"),
            password=os.getenv("DB_ADMIN", "")
        )


@dataclass
class AuthConfig:
    """Authentication configuration"""
    jwt_secret: str
    eval_user_id: str = "eval-runner"
    eval_user_email: str = "eval-runner@bigfoot-golf.com"
    token_expiry_hours: int = 1

    @classmethod
    def from_env(cls):
        return cls(
            jwt_secret=os.getenv("JWT_SECRET", ""),
            eval_user_id=os.getenv("EVAL_USER_ID", "eval-runner"),
            eval_user_email=os.getenv("EVAL_USER_EMAIL", "eval-runner@bigfoot-golf.com")
        )


@dataclass
class AWSConfig:
    """AWS Bedrock configuration"""
    region: str
    access_key_id: Optional[str]
    secret_access_key: Optional[str]
    model_id: str

    @classmethod
    def from_env(cls):
        return cls(
            region=os.getenv("AWS_REGION", "us-east-1"),
            access_key_id=os.getenv("AWS_ACCESS_KEY_ID"),
            secret_access_key=os.getenv("AWS_SECRET_ACCESS_KEY"),
            model_id=os.getenv("EVAL_MODEL_ID", "global.amazon.nova-premier-v1:0")
        )


@dataclass
class APIConfig:
    """API endpoint configuration"""
    base_url: str
    chat_endpoint: str
    mcp_endpoint: str

    @classmethod
    def from_env(cls):
        base_url = os.getenv("API_BASE_URL", "http://localhost:8000")
        return cls(
            base_url=base_url,
            chat_endpoint=f"{base_url}/api/chat",
            mcp_endpoint=os.getenv("MCP_ENDPOINT", "http://localhost:8081")
        )


@dataclass
class PhoenixConfig:
    """Phoenix observability configuration"""
    endpoint: str
    enabled: bool

    @classmethod
    def from_env(cls):
        return cls(
            endpoint=os.getenv("PHOENIX_ENDPOINT", "http://localhost:6006"),
            enabled=os.getenv("PHOENIX_ENABLED", "true").lower() == "true"
        )


@dataclass
class PathConfig:
    """File path configuration"""
    runner_dir: str
    datasets_dir: str
    results_dir: str
    evals_dir: str

    @classmethod
    def from_env(cls):
        runner_dir = os.path.dirname(os.path.abspath(__file__))
        datasets_dir = os.path.join(runner_dir, "../datasets")
        results_dir = os.path.join(runner_dir, "results")
        evals_dir = os.path.join(runner_dir, "evals")

        # Create directories if they don't exist
        os.makedirs(results_dir, exist_ok=True)
        os.makedirs(evals_dir, exist_ok=True)

        return cls(
            runner_dir=runner_dir,
            datasets_dir=datasets_dir,
            results_dir=results_dir,
            evals_dir=evals_dir
        )


@dataclass
class EvalRunnerConfig:
    """Master configuration for the evaluation runner"""
    database: DatabaseConfig
    auth: AuthConfig
    aws: AWSConfig
    api: APIConfig
    phoenix: PhoenixConfig
    paths: PathConfig

    @classmethod
    def from_env(cls):
        return cls(
            database=DatabaseConfig.from_env(),
            auth=AuthConfig.from_env(),
            aws=AWSConfig.from_env(),
            api=APIConfig.from_env(),
            phoenix=PhoenixConfig.from_env(),
            paths=PathConfig.from_env()
        )

    def validate(self) -> list[str]:
        """
        Validate configuration and return list of errors

        Returns:
            List of validation error messages
        """
        errors = []

        if not self.database.password:
            errors.append("DB_ADMIN environment variable not set")

        if not self.auth.jwt_secret:
            errors.append("JWT_SECRET environment variable not set")

        if not self.aws.region:
            errors.append("AWS_REGION environment variable not set")

        return errors
