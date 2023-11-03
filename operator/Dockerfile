FROM golang:1.18-alpine AS build-env
WORKDIR /app
COPY ./src ./
RUN go mod tidy && go build -ldflags="-s -w" -o failover-manager-operator



FROM alpine:3
WORKDIR /app

RUN addgroup app && adduser -D -G app -s /sbin/nologin app

USER app

COPY --from=build-env /app/failover-manager-operator .

ENTRYPOINT ["./failover-manager-operator"]

