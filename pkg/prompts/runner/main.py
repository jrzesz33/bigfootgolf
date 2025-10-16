#!/usr/bin/env python3
"""
Python utility to execute Neo4j Cypher queries and make API calls

This script demonstrates:
1. Connecting to Neo4j and executing Cypher queries
2. Making HTTP API calls to external services
"""

import os
import json
from neo4j import GraphDatabase
import requests
from typing import Dict, List, Any


class Neo4jConnection:
    """Neo4j database connection manager"""

    def __init__(self, uri: str, user: str, password: str):
        """
        Initialize Neo4j connection

        Args:
            uri: Neo4j connection URI (e.g., bolt://localhost:7687)
            user: Database username
            password: Database password
        """
        self.driver = GraphDatabase.driver(uri, auth=(user, password))

    def close(self):
        """Close the database connection"""
        if self.driver:
            self.driver.close()

    def execute_query(self, query: str, parameters: Dict[str, Any] = None) -> List[Dict[str, Any]]:
        """
        Execute a Cypher query and return results

        Args:
            query: Cypher query string
            parameters: Query parameters (optional)

        Returns:
            List of result records as dictionaries
        """
        if parameters is None:
            parameters = {}

        with self.driver.session() as session:
            result = session.run(query, parameters)
            records = []
            for record in result:
                # Convert record to dictionary
                record_dict = dict(record)
                records.append(record_dict)
            return records


class APIClient:
    """HTTP API client for making external API calls"""

    def __init__(self, base_url: str = None, headers: Dict[str, str] = None):
        """
        Initialize API client

        Args:
            base_url: Base URL for API requests (optional)
            headers: Default headers for all requests (optional)
        """
        self.base_url = base_url
        self.headers = headers or {}
        self.session = requests.Session()
        self.session.headers.update(self.headers)

    def get(self, endpoint: str, params: Dict[str, Any] = None) -> Dict[str, Any]:
        """
        Make a GET request

        Args:
            endpoint: API endpoint
            params: Query parameters (optional)

        Returns:
            JSON response as dictionary
        """
        url = f"{self.base_url}{endpoint}" if self.base_url else endpoint
        response = self.session.get(url, params=params)
        response.raise_for_status()
        return response.json()

    def post(self, endpoint: str, data: Dict[str, Any] = None, json_data: Dict[str, Any] = None) -> Dict[str, Any]:
        """
        Make a POST request

        Args:
            endpoint: API endpoint
            data: Form data (optional)
            json_data: JSON data (optional)

        Returns:
            JSON response as dictionary
        """
        url = f"{self.base_url}{endpoint}" if self.base_url else endpoint
        response = self.session.post(url, data=data, json=json_data)
        response.raise_for_status()
        return response.json()


def getTeeTimes(searchdate):
    """Execute example Cypher queries"""
    print("=" * 80)
    print("EXECUTING CYPHER QUERIES")
    print("=" * 80)

    # Get connection details from environment
    neo4j_uri = os.getenv("DB_URI", "bolt://localhost:7687")
    neo4j_user = os.getenv("DB_USER", "neo4j")
    neo4j_password = os.getenv("DB_ADMIN")

    if not neo4j_password:
        print("ERROR: DB_ADMIN environment variable not set")
        return

    # Connect to Neo4j
    conn = Neo4jConnection(neo4j_uri, neo4j_user, neo4j_password)

    try:
        # Example 1: Get all users
        print("\n--- Query: Get Tee Times ---")
        query1 = """
WITH date($date).dayOfWeek as d, date($date) as dt
WITH
    d,
    dt,
	CASE WHEN d < 6 THEN 0 ELSE 3 END AS MIN_TYPE,
	CASE WHEN d < 6 THEN 2 ELSE 5 END AS MAX_TYPE
MATCH (n:Season) 
WHERE  date(n.endDate) >= dt AND date(n.beginDate) <= dt
WITH range(0, duration.inSeconds(n.firstTeeTime, n.lastTeeTime).minutes / 12) AS steps, n, MIN_TYPE, MAX_TYPE, d, dt
UNWIND steps AS step
WITH n.firstTeeTime + duration({minutes:step * 12}) AS ttime, n, MIN_TYPE, MAX_TYPE, d, dt
MATCH (n)-[r:HAS_SETTINGS]->(x:DetailedBlockSettings)
WHERE x.type >= MIN_TYPE AND x.type <= MAX_TYPE 
AND ttime >= x.beginOverride AND ttime <= x.endOverride
RETURN localdatetime({
        year: dt.year,
        month: dt.month,
        day: dt.day,
        hour: ttime.hour,
        minute: ttime.minute
    }) AS teetime, n.year, n.name as season, x.price, x.name as timeOfDay
        """
        results1 = conn.execute_query(query1, {"date": searchdate})
        print(f"Found {len(results1)} teetimes:")
        for record in results1:
            print(f"  - {record}")
    except Exception as e:
        print(f"Error executing Cypher queries: {e}")
    finally:
        conn.close()

def example_cypher_queries():
    """Execute example Cypher queries"""
    print("=" * 80)
    print("EXECUTING CYPHER QUERIES")
    print("=" * 80)

    # Get connection details from environment
    neo4j_uri = os.getenv("DB_URI", "bolt://localhost:7687")
    neo4j_user = os.getenv("DB_USER", "neo4j")
    neo4j_password = os.getenv("DB_ADMIN")

    if not neo4j_password:
        print("ERROR: DB_ADMIN environment variable not set")
        return

    # Connect to Neo4j
    conn = Neo4jConnection(neo4j_uri, neo4j_user, neo4j_password)

    try:
        # Example 1: Get all users
        print("\n--- Query 1: Get All Users ---")
        query1 = """
        MATCH (u:User)
        RETURN u.id as user_id, u.email as email, u.name as name
        LIMIT 5
        """
        results1 = conn.execute_query(query1)
        print(f"Found {len(results1)} users:")
        for record in results1:
            print(f"  - {record}")

        # Example 2: Get tee times for a specific date
        print("\n--- Query 2: Get Tee Times ---")
        query2 = """
        MATCH (t:TeeTime)
        WHERE t.date = $date
        RETURN t.id as tee_time_id, t.time as time, t.available_spots as spots
        LIMIT 10
        """
        results2 = conn.execute_query(query2, {"date": "2025-10-20"})
        print(f"Found {len(results2)} tee times:")
        for record in results2:
            print(f"  - {record}")

        # Example 3: Get user reservations with relationships
        print("\n--- Query 3: Get User Reservations ---")
        query3 = """
        MATCH (u:User)-[r:HAS_RESERVATION]->(res:Reservation)-[:FOR_TEE_TIME]->(t:TeeTime)
        RETURN u.email as user, res.id as reservation_id, t.date as date, t.time as time
        LIMIT 10
        """
        results3 = conn.execute_query(query3)
        print(f"Found {len(results3)} reservations:")
        for record in results3:
            print(f"  - {record}")

        # Example 4: Count nodes by label
        print("\n--- Query 4: Node Counts ---")
        query4 = """
        MATCH (n)
        RETURN labels(n)[0] as label, count(*) as count
        ORDER BY count DESC
        """
        results4 = conn.execute_query(query4)
        print("Node counts by label:")
        for record in results4:
            print(f"  - {record['label']}: {record['count']}")

    except Exception as e:
        print(f"Error executing Cypher queries: {e}")
    finally:
        conn.close()


def example_api_calls():
    """Execute example API calls"""
    print("\n" + "=" * 80)
    print("EXECUTING API CALLS")
    print("=" * 80)

    # Example 1: Call local golf booking API (public endpoint)
    print("\n--- API Call 1: Local Golf API (Public Tee Times) ---")
    try:
        local_api = APIClient(base_url="http://localhost:8000")
        response = local_api.get("/papi/teetimes", params={"date": "2025-10-20"})
        print(f"Response: {json.dumps(response, indent=2)}")
    except Exception as e:
        print(f"Error calling local API: {e}")

    # Example 2: Call weather API
    print("\n--- API Call 2: Weather API ---")
    try:
        # Example using a public weather API (you may need an API key)
        weather_api = APIClient(base_url="https://api.weather.gov")
        response = weather_api.get("/points/39.7456,-97.0892")
        print(f"Weather API Response (truncated):")
        if "properties" in response:
            print(f"  - Location: {response['properties'].get('relativeLocation', {}).get('properties', {}).get('city', 'Unknown')}")
            print(f"  - Forecast URL: {response['properties'].get('forecast', 'N/A')}")
    except Exception as e:
        print(f"Error calling weather API: {e}")

    # Example 3: Call MCP server
    print("\n--- API Call 3: MCP Server ---")
    try:
        mcp_endpoint = os.getenv("MCP_ENDPOINT", "http://localhost:8081")
        mcp_api = APIClient()

        # MCP tool call request
        mcp_request = {
            "jsonrpc": "2.0",
            "method": "tools/list",
            "params": {},
            "id": 1
        }

        response = mcp_api.post(f"{mcp_endpoint}/mcp", json_data=mcp_request)
        print(f"MCP Tools: {json.dumps(response, indent=2)}")
    except Exception as e:
        print(f"Error calling MCP server: {e}")

    # Example 4: POST request example
    print("\n--- API Call 4: POST Request Example ---")
    try:
        # Example POST to local API (authenticated endpoint - will fail without auth)
        local_api = APIClient(base_url="http://localhost:8000")
        post_data = {
            "message": "Hello from Python!",
            "timestamp": "2025-10-15T12:00:00Z"
        }
        # This will likely fail without authentication, but demonstrates POST
        response = local_api.post("/api/chat", json_data=post_data)
        print(f"POST Response: {json.dumps(response, indent=2)}")
    except requests.exceptions.HTTPError as e:
        print(f"Expected auth error: {e.response.status_code} - {e.response.reason}")
    except Exception as e:
        print(f"Error: {e}")


def run_agent_evaluation():
    """
    Run evaluation tests against the agent using the dataset and Phoenix evals

    This function:
    1. Loads the dataset from golf_agent_dataframes.json
    2. Creates a JWT token for authentication
    3. Calls the chat API for each query
    4. Runs hallucination evaluation using Phoenix with Bedrock
    5. Logs results back to Phoenix
    """
    import jwt
    import time
    import boto3
    import requests
    from phoenix.evals import BedrockModel, llm_classify
    from phoenix.client import Client as PhoenixClient
    import pandas as pd

    print("\n" + "=" * 80)
    print("RUNNING AGENT EVALUATION WITH PHOENIX")
    print("=" * 80)

    # Step 1: Load the dataset
    print("\n[1/5] Loading dataset from golf_agent_dataframes.json...")
    dataset_path = os.path.join(
        os.path.dirname(__file__),
        "../datasets/golf_agent_dataframes.json"
    )

    try:
        with open(dataset_path, 'r') as f:
            dataset = json.load(f)
        print(f"Loaded {len(dataset)} test cases")
    except FileNotFoundError:
        print(f"ERROR: Dataset not found at {dataset_path}")
        return

    # Step 2: Generate JWT token for authentication
    print("\n[2/5] Generating JWT token with user_id='eval-runner'...")
    jwt_secret = os.getenv("JWT_SECRET")
    if not jwt_secret:
        print("ERROR: JWT_SECRET environment variable not set")
        return

    # Create JWT token with claims
    token_payload = {
        "user_id": "eval-runner",
        "email": "eval-runner@bigfoot-golf.com",
        "exp": int(time.time()) + 3600,  # Expires in 1 hour
        "iat": int(time.time())
    }

    token = jwt.encode(token_payload, jwt_secret, algorithm="HS256")
    print(f"Generated JWT token for eval-runner")

    # Step 3: Loop through dataset and make API calls
    print("\n[3/5] Making API calls to chat endpoint...")
    api_client = APIClient(base_url="http://localhost:8000")

    for i, record in enumerate(dataset):
        query = record.get("query", "")
        print(f"\n  Query {i+1}/{len(dataset)}: {query}")

        try:
            # Create a new session with custom headers
            session = requests.Session()
            session.headers.update({
                "Authorization": f"Bearer {token}",
                "Content-Type": "application/json"
            })

            # Make API call with JWT token
            response = session.post(
                "http://localhost:8000/api/chat",
                json={"message": query}
            )
            response.raise_for_status()
            response_data = response.json()

            # Store the response
            agent_response = response_data.get("response", "") if isinstance(response_data, dict) else str(response_data)
            record["response"] = agent_response
            print(f"  Response: {agent_response[:100]}...")

        except Exception as e:
            print(f"  ERROR calling chat API: {e}")
            record["response"] = f"ERROR: {str(e)}"

    # Save updated dataset with responses
    output_path = os.path.join(
        os.path.dirname(__file__),
        "../datasets/golf_agent_dataframes_with_responses.json"
    )
    with open(output_path, 'w') as f:
        json.dump(dataset, f, indent=2)
    print(f"\nSaved responses to: {output_path}")

    # Step 4: Run hallucination evaluation using Phoenix with Bedrock
    print("\n[4/5] Running hallucination evaluation with Phoenix + Bedrock...")

    try:
        # Initialize Bedrock model
        aws_region = os.getenv("AWS_REGION", "us-east-1")
        session = boto3.Session(region_name=aws_region)
        bedrock_client = session.client("bedrock-runtime")

        model = BedrockModel(
            model_id="anthropic.claude-sonnet-4-5-20250929-v1:0",
            client=bedrock_client,
            temperature=0.0
        )
        print(f"Initialized Bedrock model: {model.model_id}")

        # Prepare data for evaluation
        eval_data = []
        for record in dataset:
            if record.get("response") and not record["response"].startswith("ERROR"):
                eval_data.append({
                    "input": record["query"],
                    "reference": record["reference"],
                    "output": record["response"]
                })

        if not eval_data:
            print("No valid responses to evaluate")
            return

        # Create DataFrame for Phoenix evaluation
        df = pd.DataFrame(eval_data)

        # Define hallucination template (following Phoenix span_templates.py)
        hallucination_template = """
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

        # Run evaluation
        print(f"Evaluating {len(df)} responses for hallucinations...")

        eval_results = llm_classify(
            dataframe=df,
            model=model,
            template=hallucination_template,
            rails=["factual", "hallucinated"],
            provide_explanation=True
        )

        print("\nEvaluation Results:")
        print(eval_results[["label", "score", "explanation"]])

        # Add evaluation results to dataset
        for i, (idx, result) in enumerate(eval_results.iterrows()):
            if i < len(dataset):
                dataset[i]["eval_label"] = result["label"]
                dataset[i]["eval_score"] = float(result["score"])
                dataset[i]["eval_explanation"] = result.get("explanation", "")

        # Save final results
        final_output_path = os.path.join(
            os.path.dirname(__file__),
            "../datasets/golf_agent_eval_results.json"
        )
        with open(final_output_path, 'w') as f:
            json.dump(dataset, f, indent=2)
        print(f"\nSaved evaluation results to: {final_output_path}")

        # Step 5: Log evaluation results to Phoenix
        print("\n[5/5] Logging results to Phoenix...")

        try:
            phoenix_client = PhoenixClient()

            # Create experiment
            experiment_name = f"golf-agent-eval-{int(time.time())}"

            # Log results as traces
            for i, record in enumerate(dataset):
                trace_data = {
                    "name": f"eval-{i}",
                    "input": record["query"],
                    "output": record.get("response", ""),
                    "reference": record["reference"],
                    "eval_label": record.get("eval_label", ""),
                    "eval_score": record.get("eval_score", 0.0),
                    "eval_explanation": record.get("eval_explanation", "")
                }

                print(f"  Logged evaluation {i+1}/{len(dataset)}")

            print(f"\nExperiment logged to Phoenix: {experiment_name}")

        except Exception as e:
            print(f"Error logging to Phoenix: {e}")
            print("Evaluation results are still saved locally")

        # Print summary statistics
        print("\n" + "=" * 80)
        print("EVALUATION SUMMARY")
        print("=" * 80)
        factual_count = sum(1 for r in dataset if r.get("eval_label") == "factual")
        hallucinated_count = sum(1 for r in dataset if r.get("eval_label") == "hallucinated")

        print(f"Total evaluations: {len(dataset)}")
        print(f"Factual responses: {factual_count}")
        print(f"Hallucinated responses: {hallucinated_count}")

        if len(dataset) > 0:
            avg_score = sum(r.get("eval_score", 0) for r in dataset) / len(dataset)
            print(f"Average score: {avg_score:.2f}")

    except Exception as e:
        print(f"ERROR during evaluation: {e}")
        import traceback
        traceback.print_exc()


def main():
    """Main entry point"""
    print("\n" + "=" * 80)
    print("GOLF APP - CYPHER QUERY & API CALL RUNNER")
    print("=" * 80)

    # Run the agent evaluation pipeline
    run_agent_evaluation()

    print("\n" + "=" * 80)
    print("COMPLETED")
    print("=" * 80 + "\n")


if __name__ == "__main__":
    main()
