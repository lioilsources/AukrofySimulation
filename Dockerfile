# Build stage — statická binárka bez CGO (modernc.org/sqlite je pure Go)
FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/engine ./cmd/engine \
 && CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/reporter ./cmd/reporter

# Runtime stage
FROM alpine:3.21
RUN apk add --no-cache ca-certificates wget tzdata \
 && addgroup -S app && adduser -S app -G app
WORKDIR /app
COPY --from=build /out/engine /out/reporter /app/
COPY prompts/ /app/prompts/
COPY web/ /app/web/
RUN mkdir -p /app/data /app/reports && chown -R app:app /app
USER app

EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD wget -qO- http://localhost:8080/healthz || exit 1

CMD ["/app/engine"]
