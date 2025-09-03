check_and_exit() {
  local var_name="$1"
  if [[ -z "${!var_name}" ]]; then
    echo "Error: Environment variable '$var_name' is not set." >&2
    exit 1
  fi
}

check_and_exit "DB_ADMIN"

docker run -d \
    --restart always \
    --publish=7474:7474 --publish=7687:7687 \
    --env NEO4J_AUTH=neo4j/$DB_ADMIN \
    --name neo_bird_db \
    neo4j:2025.05.0

# ADD THIS LATER --volume=/WHERE/db/data:/data \