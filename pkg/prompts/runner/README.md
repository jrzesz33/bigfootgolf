# Golf Agent Evaluation Runner

Python utility to execute Neo4j Cypher queries, make HTTP API calls, and run comprehensive agent evaluations using Phoenix and AWS Bedrock.

## Features

- **Neo4j Cypher Queries**: Execute queries against the golf booking Neo4j database
- **HTTP API Calls**: Make GET/POST requests to various APIs (local golf API, weather API, MCP server)
- **Agent Evaluation Pipeline**: Complete evaluation system with:
  - Dataset loading from `golf_agent_dataframes.json`
  - JWT token generation for authenticated API calls
  - Automated chat API testing
  - Hallucination detection using Phoenix Evals + AWS Bedrock
  - Results logging to Phoenix observability platform
- **Reusable Classes**: `Neo4jConnection` and `APIClient` classes for easy integration

## Installation

```bash
cd pkg/prompts/runner
pip install -r requirements.txt
```

## Environment Variables

Create a `.env` file or set these environment variables (see `.env.example`):

```bash
# Required for Neo4j
export DB_ADMIN="your-neo4j-password"
export DB_URI="bolt://localhost:7687"
export DB_USER="neo4j"

# Required for JWT token generation
export JWT_SECRET="your-jwt-secret"

# Required for AWS Bedrock evaluations
export AWS_REGION="us-east-1"
export AWS_ACCESS_KEY_ID="your-aws-access-key"
export AWS_SECRET_ACCESS_KEY="your-aws-secret-key"

# Optional (with defaults)
export MCP_ENDPOINT="http://localhost:8081"
export PHOENIX_ENDPOINT="http://localhost:6006"
```

## Usage

### Run Agent Evaluation Pipeline

```bash
python main.py
```

This will execute the complete evaluation pipeline:

1. **Load Dataset**: Reads test cases from `golf_agent_dataframes.json`
2. **Generate JWT Token**: Creates authentication token with `user_id="eval-runner"`
3. **Call Chat API**: Sends each query to `/api/chat` endpoint
4. **Run Phoenix Evals**: Uses AWS Bedrock to evaluate responses for hallucinations
5. **Log Results**: Saves results locally and logs to Phoenix

### Output Files

The evaluation pipeline generates these files in `/pkg/prompts/datasets/`:

- `golf_agent_dataframes_with_responses.json` - Dataset with agent responses
- `golf_agent_eval_results.json` - Complete evaluation results with scores

### Example Output

```
================================================================================
GOLF APP - CYPHER QUERY & API CALL RUNNER
================================================================================

--- Query 1: Get All Users ---
Found 5 users:
  - {'user_id': '123', 'email': 'user@example.com', 'name': 'John Doe'}
  ...

--- Query 2: Get Tee Times ---
Found 10 tee times:
  - {'tee_time_id': 'tt1', 'time': '08:00', 'spots': 4}
  ...

--- API Call 1: Local Golf API (Public Tee Times) ---
Response: {...}
```

## Code Examples

### Using Neo4jConnection

```python
from main import Neo4jConnection

conn = Neo4jConnection(
    uri="bolt://localhost:7687",
    user="neo4j",
    password="your-password"
)

# Execute query
query = "MATCH (u:User) RETURN u.email as email LIMIT 10"
results = conn.execute_query(query)

for record in results:
    print(record['email'])

conn.close()
```

### Using APIClient

```python
from main import APIClient

# Create client with base URL
api = APIClient(base_url="http://localhost:8000")

# GET request
tee_times = api.get("/papi/teetimes", params={"date": "2025-10-20"})
print(tee_times)

# POST request
response = api.post("/api/endpoint", json_data={"key": "value"})
print(response)
```

## Example Queries Included

### 1. Get All Users
```cypher
MATCH (u:User)
RETURN u.id, u.email, u.name
LIMIT 5
```

### 2. Get Tee Times for Date
```cypher
MATCH (t:TeeTime)
WHERE t.date = $date
RETURN t.id, t.time, t.available_spots
```

### 3. Get User Reservations with Relationships
```cypher
MATCH (u:User)-[r:HAS_RESERVATION]->(res:Reservation)-[:FOR_TEE_TIME]->(t:TeeTime)
RETURN u.email, res.id, t.date, t.time
```

### 4. Count Nodes by Label
```cypher
MATCH (n)
RETURN labels(n)[0] as label, count(*) as count
ORDER BY count DESC
```

## API Endpoints Tested

1. **Local Golf API**: `http://localhost:8000/papi/teetimes`
2. **Weather API**: `https://api.weather.gov`
3. **MCP Server**: `http://localhost:8081/mcp`
4. **POST Example**: `http://localhost:8000/api/chat`

## Troubleshooting

### Neo4j Connection Issues

Ensure Neo4j is running:
```bash
docker-compose up -d neo4j
```

Verify connection:
```bash
echo $DB_ADMIN  # Should output your password
```

### API Connection Issues

Ensure the golf app server is running:
```bash
cd /workspaces/golf_app
MODE=dev go run web/main.go
```

Ensure MCP server is running:
```bash
cd mcp
go run . -mode server
```

### Import Errors

If you get import errors:
```bash
pip install --upgrade neo4j requests
```

## Evaluation Pipeline Details

### 1. Dataset Structure

The `golf_agent_dataframes.json` file contains test cases with:

```json
[
  {
    "reference": "Expected behavior or facts",
    "query": "User question to test",
    "response": "" // Filled in by the runner
  }
]
```

### 2. JWT Token Generation

The runner creates a JWT token matching your Go application's format:

```python
token_payload = {
    "user_id": "eval-runner",
    "email": "eval-runner@bigfoot-golf.com",
    "exp": int(time.time()) + 3600,
    "iat": int(time.time())
}
token = jwt.encode(token_payload, jwt_secret, algorithm="HS256")
```

### 3. Hallucination Evaluation

Uses Phoenix's hallucination template to compare:
- **Input**: User query
- **Reference**: Expected behavior/facts
- **Output**: Agent's actual response

Results in classification:
- `factual` (score: 1.0) - Response supported by reference
- `hallucinated` (score: 0.0) - Response contains unsupported info

### 4. Bedrock Model Configuration

The evaluation uses AWS Bedrock with Claude Sonnet 4.5:

```python
model = BedrockModel(
    model_id="anthropic.claude-sonnet-4-5-20250929-v1:0",
    client=bedrock_client,
    temperature=0.0
)
```

## Prerequisites

Before running the evaluation:

1. **Golf App Server Running**: `cd /workspaces/golf_app && MODE=dev go run web/main.go`
2. **Neo4j Running**: `docker-compose up -d neo4j`
3. **AWS Credentials Configured**: Set AWS environment variables or use `~/.aws/credentials`
4. **Phoenix Running** (optional): For results logging

## Troubleshooting

### Authentication Errors

If you get 401/403 errors, verify:
```bash
echo $JWT_SECRET  # Should match your Go app's JWT_SECRET
```

### AWS Bedrock Errors

Ensure you have:
- Valid AWS credentials with Bedrock access
- Proper IAM permissions for `bedrock-runtime:InvokeModel`
- Requested access to Claude models in your AWS region

### Dataset Not Found

Ensure the dataset exists:
```bash
ls -la /workspaces/golf_app/pkg/prompts/datasets/golf_agent_dataframes.json
```

## Integration with Other Tools

This runner integrates with:
- **Evaluation System** (`/pkg/prompts/evals/`) - Core eval logic
- **A2A Agent** (`/pkg/prompts/agent/`) - Can be tested via this runner
- **Dataset Generation** (`/pkg/prompts/datasets/`) - Source of test cases
- **Phoenix Observability** - Results logging and tracking

## Example Evaluation Output

```
================================================================================
RUNNING AGENT EVALUATION WITH PHOENIX
================================================================================

[1/5] Loading dataset from golf_agent_dataframes.json...
Loaded 3 test cases

[2/5] Generating JWT token with user_id='eval-runner'...
Generated JWT token for eval-runner

[3/5] Making API calls to chat endpoint...
  Query 1/3: What is the weather at the golf course over the weekend?
  Response: Based on the weather forecast...

[4/5] Running hallucination evaluation with Phoenix + Bedrock...
Initialized Bedrock model: anthropic.claude-sonnet-4-5-20250929-v1:0
Evaluating 3 responses for hallucinations...

Evaluation Results:
   label        score  explanation
0  factual      1.0    Response correctly references weather API
1  factual      1.0    Response provides appropriate tee time list
2  factual      1.0    Response correctly limits to Bigfoot Golf Course

[5/5] Logging results to Phoenix...
Experiment logged to Phoenix: golf-agent-eval-1234567890

================================================================================
EVALUATION SUMMARY
================================================================================
Total evaluations: 3
Factual responses: 3
Hallucinated responses: 0
Average score: 1.00
```
