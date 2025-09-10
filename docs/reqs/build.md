# Build Components 

The Build Component should be executed when a Pull Request within the repository is approved and merged back into the Main branch.

## REQUIREMENTS:
1. Build the WASM Application
   - Compile the WASM Application into the public folder where it will be served as an asset

2. Build the Go Lang Web Application
   - Compile the Go Binary to run within a linux container and be served from an AWS Service

3. Build the Docker Containers
   - Build the Neo4j Database from the Dockerfile.db file
   - Build the Web Application Image from the Dockerfile.webapp file
   - Tag and Push the images to the Container Registry

## TECHNICAL PREFERENCES:
- Github Actions should be used as the Orchestration mechanism
- Github Container Registry should be used as the Artifact Registry