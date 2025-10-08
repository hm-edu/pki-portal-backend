FROM golang:1.25.2-bookworm AS builder

ARG SERVICE
RUN apt update && apt install -y ca-certificates wget git && update-ca-certificates
COPY backend/common /app/backend/common
COPY backend/${SERVICE} /app/backend/${SERVICE}
WORKDIR /app/backend/${SERVICE}
RUN CGO_ENABLED=1 GOOS=linux go build -a -ldflags '-linkmode external -extldflags "-static"' -tags sqlite_omit_load_extension -o /app/service .
# Adding the grpc_health_probe
RUN GRPC_HEALTH_PROBE_VERSION=v0.4.1 && \
    wget -qO/bin/grpc_health_probe https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/${GRPC_HEALTH_PROBE_VERSION}/grpc_health_probe-linux-amd64 && \
    chmod +x /bin/grpc_health_probe

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/service /entrypoint
COPY --from=builder /bin/grpc_health_probe ./grpc_health_probe
ENTRYPOINT ["/entrypoint", "run"]
