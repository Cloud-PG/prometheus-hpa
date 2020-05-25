FROM ubuntu

# Install prerequisites
RUN apt-get update && apt-get install -y \
curl

COPY ./process_exporter /bin/process-exporter

# Run the process-exporter command by default when the container starts.
ENTRYPOINT ["/bin/process-exporter"]
