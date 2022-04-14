FROM golang:alpine as builder
ARG SERVICE 
RUN apk --no-cache add ca-certificates
COPY backend/common /app/backend/common
COPY /backend/${SERVICE} /app/backend/${SERVICE}
WORKDIR /app/backend/${SERVICE}
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /app/service . 

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/service /entrypoint
ENTRYPOINT ["/entrypoint", "run"]