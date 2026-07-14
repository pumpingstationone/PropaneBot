# --- Build stage ---
FROM golang:1.25-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/propanebot .

# --- Runtime stage ---
FROM alpine:3.20

# ca-certificates: outbound TLS to Discord/MQTT
# tzdata: mqtt.go localizes timestamps to America/Chicago
RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S propanebot && adduser -S propanebot -G propanebot

WORKDIR /app

COPY --from=builder /out/propanebot ./propanebot
# Seed a default cylinder.json; mount your own over it (see README) to persist edits
# made through the /cylinder web page across container recreation.
COPY --from=builder /src/cylinder.json ./cylinder.json

RUN chown -R propanebot:propanebot /app
USER propanebot

# config.json is not baked into the image since it holds MQTT/Discord/Slack
# secrets - it must be bind-mounted in at runtime.
EXPOSE 9991

ENTRYPOINT ["./propanebot"]
