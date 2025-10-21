# Dockerfile for agentgateway
# This uses the official agentgateway image or builds from binary

FROM ghcr.io/agentgateway/agentgateway:latest
# Set working directory
WORKDIR /app

# Copy the config file
COPY config.yaml /app/config.yaml

# Expose ports defined in config (3000 and 3001)
EXPOSE 3000 3001 3002 15000

# Run agentgateway with the config file
ENTRYPOINT ["/app/agentgateway"]
CMD ["-f", "/app/config.yaml"]
