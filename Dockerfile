FROM golang:alpine as builder
ARG SERVICE 
RUN apk --no-cache add ca-certificates gcc musl-dev
COPY backend/common /app/backend/common
COPY /backend/${SERVICE} /app/backend/${SERVICE}
WORKDIR /app/backend/${SERVICE}
RUN CGO_ENABLED=1 GOOS=linux go build -a -ldflags '-linkmode external -extldflags "-static"' -tags sqlite_omit_load_extension -o /app/service . 

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/service /entrypoint
ENTRYPOINT ["/entrypoint", "run"]