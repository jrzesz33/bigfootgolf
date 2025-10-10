| Type | Metric | Description | Possible Tooling |
| :---- | :---- | :---- | :---- |
| **Drift Tolerance** | Statistical Drift Scores (KS, JSD, PSI…) | Quantifies distribution changes between training and production data using Kolmogorov-Smirnov, Jensen-Shannon Divergence, Population Stability Index | Arize AI, Fiddler AI, WhyLabs, Custom Python (scipy, numpy) |
|  | Embedding | Monitors semantic drift in vector representations of inputs/outputs | LangSmith, Weights & Biases, Custom vector similarity tracking |
|  | Feature | Tracks changes in input feature distributions over time | Evidently AI, Great Expectations, Arize |
|  | Prediction | Monitors changes in model output distributions | Arize AI, Fiddler, WhyLabs |
|  | Concept | Detects when the relationship between inputs and outputs changes | NannyML, Alibi Detect, Custom ML monitoring |
|  | Perplexity | Measures model uncertainty/confidence degradation over time | LangSmith, Custom LLM logging, Weights & Biases |
| **Model Accuracy** | Response Relevance | Evaluates how well responses address the user's query | RAGAS, LangSmith, Trulens, Phoenix Arize |
|  | Hallucination Rate | Percentage of responses containing false or ungrounded information | Galileo, Patronus AI, RAGAS, Custom LLM-as-judge |
|  | Task Success Rate | Percentage of agent tasks completed successfully | LangSmith, Custom telemetry, Datadog |
|  | User Satisfaction | Explicit or implicit user feedback on response quality | Qualtrics, Custom feedback loops, Thumbs up/down tracking |
|  | Avg. Steps | Mean number of steps agent takes to complete tasks | LangGraph Studio, LangSmith, Custom observability |
|  | Tool Success | Rate of successful tool/function call executions | LangSmith, OpenTelemetry, Custom logging |
| **Detection Accuracy** | Precision/Recall/F1 from Metrics | Standard classification metrics for detection tasks | scikit-learn, Custom evaluation pipelines, MLflow |
| **RAG Performance** | Context Relevance | Measures how relevant retrieved documents are to the query | RAGAS, Trulens, LlamaIndex, LangSmith |
|  | Groundedness | Evaluates if response claims are supported by retrieved context | Patronus AI, RAGAS, Galileo, Custom verification |
|  | Avg Doc Retrieval | Mean number of documents retrieved per query | Pinecone Analytics, Weaviate metrics, Custom logging |
| **Guardrails** | Content Filter Result Ratio | Rate at which safety filters trigger (blocks/warnings/passes) | Azure AI Content Safety, Guardrails AI, LlamaGuard, Lakera Guard |
| **Operational** | Token Per Second | Throughput metric for inference speed | LangSmith, OpenAI dashboard, Custom timing |
|  | Cache Hit Ratio | Percentage of requests served from cache vs fresh generation | Redis metrics, Custom caching layer, Prompt caching analytics |
|  | Latency | End-to-end response time (p50, p95, p99) | Datadog, New Relic, OpenTelemetry, LangSmith |
|  | Token Cost | Per-request and aggregate token expenditure | OpenAI usage dashboard, Custom cost tracking, LangSmith |
|  | Error Rate | Percentage of requests resulting in errors/failures | Sentry, Datadog, Custom error logging |
| **Reasoning** | Confidence Score | Model's self-assessed certainty in its responses | Custom prompting for confidence, Patronus AI, LLM introspection |
|  | Conclusion Ratio | Rate at which agent reaches definitive conclusions vs uncertain outcomes | Custom evaluation, LangSmith traces, Agent-specific metrics |