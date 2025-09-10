#!/bin/bash
set -e

# Override Neo4j password if NEO4J_PASSWORD is set
if [ -n "$NEO4J_PASSWORD" ]; then
    export NEO4J_AUTH="neo4j/$NEO4J_PASSWORD"
fi

# Function to wait for Neo4j to be ready
wait_for_neo4j() {
    echo "Waiting for Neo4j to be ready..."
    until cypher-shell -a bolt://localhost:7687 -u neo4j -p "${NEO4J_PASSWORD:-changeme}" "RETURN 1;" >/dev/null 2>&1; do
        sleep 2
    done
    echo "Neo4j is ready!"
}

# Function to run bootstrap scripts
run_bootstrap_scripts() {
    echo "Running bootstrap scripts..."
    
    # Run Cypher scripts
    for script in /docker-entrypoint-initdb.d/*.cypher; do
        if [ -r "$script" ]; then
            echo "Running $script"
            cypher-shell -a bolt://localhost:7687 -u neo4j -p "${NEO4J_PASSWORD:-changeme}" --file "$script"
        fi
    done
    
    echo "Bootstrap completed!"
}

# Start Neo4j in the background
echo "Starting Neo4j..."
/docker-entrypoint.sh neo4j &
NEO4J_PID=$!

# Wait for Neo4j to be ready, then run bootstrap scripts
if [ "$(ls -A /docker-entrypoint-initdb.d/ 2>/dev/null)" ]; then
    wait_for_neo4j
    run_bootstrap_scripts
fi

# Wait for Neo4j process to complete
wait $NEO4J_PID
