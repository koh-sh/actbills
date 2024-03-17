FROM golang:1.21-alpine AS build-stage
WORKDIR /app
COPY . /app/

RUN CGO_ENABLED=0 GOOS=linux go build -o /actbills

FROM alpine:latest AS build-release-stage

WORKDIR /

COPY --from=build-stage /actbills /usr/local/bin/actbills

COPY scripts/entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
