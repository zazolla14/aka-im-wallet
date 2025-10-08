# Multi-stage build: Install dlv directly from golang image
FROM ran7891/alpine:1.25.1-dlv

# Build arguments for service customization
ARG SERVICE_NAME
ARG BINARY_PATH=bundle

# Ensure dlv is executable
RUN chmod +x /usr/local/bin/dlv

# Set working directory
WORKDIR /app

# Copy binary files (copy from current build directory)
COPY ${BINARY_PATH}/${SERVICE_NAME} .

# Expose debug port
EXPOSE 2345

RUN echo "# Debug Instructions" > /app/DEBUGME.md && \
    echo "" >> /app/DEBUGME.md && \
    echo "=== Debug Run ===" >> /app/DEBUGME.md && \
    echo "" >> /app/DEBUGME.md && \
    echo "dlv --listen=:2345 --headless=true --api-version=2 --accept-multiclient exec /app/${SERVICE_NAME} -- -c /config" >> /app/DEBUGME.md && \
    echo "" >> /app/DEBUGME.md && \
    echo "Connection Info:" >> /app/DEBUGME.md && \
    echo "- Debug Port: 2345" >> /app/DEBUGME.md && \
    echo "- API Version: 2" >> /app/DEBUGME.md && \
    echo "- Accept Multi-client: Yes" >> /app/DEBUGME.md && \
    echo "- Binary: /app/${SERVICE_NAME}" >> /app/DEBUGME.md && \
    echo "- Config: /config" >> /app/DEBUGME.md

# Create entrypoint script to handle dynamic service name with dlv
RUN echo "#!/bin/sh" > /app/entrypoint.sh && \
    echo "echo \"=== Binary Information ===\"" >> /app/entrypoint.sh && \
    echo "echo \"Service: ${SERVICE_NAME}\"" >> /app/entrypoint.sh && \
    echo "echo \"Binary Path: /app/${SERVICE_NAME}\"" >> /app/entrypoint.sh && \
    echo "echo \"Binary MD5: \$(md5sum /app/${SERVICE_NAME} | cut -d' ' -f1)\"" >> /app/entrypoint.sh && \
    echo "echo \"Build Time: \$(date)\"" >> /app/entrypoint.sh && \
    echo "echo \"Debug Port: 2345 (connect anytime)\"" >> /app/entrypoint.sh && \
    echo "echo \"======================== ===\"" >> /app/entrypoint.sh && \
    echo "/app/${SERVICE_NAME} -c /config" >> /app/entrypoint.sh && \
    chmod +x /app/entrypoint.sh

# Default entrypoint (using dynamic service name)
CMD ["/app/entrypoint.sh"]
