# ==========================================
# STAGE 1: BUILD GO APPLICATION
# ==========================================
FROM golang:1.25-alpine AS builder

WORKDIR /src

RUN apk add --no-cache git ca-certificates tzdata

# Copy Go dependency files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy backend source code
COPY . .

# Build Go binary
RUN CGO_ENABLED=0 GOOS=linux \
    go build -trimpath \
    -ldflags="-s -w" \
    -o /out/helpdesk-server .

# ==========================================
# STAGE 2: RUNTIME
# ==========================================
FROM alpine:3.23

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

# Create application user
RUN addgroup -S -g 10001 app && \
    adduser -S -u 10001 -G app app

# Copy compiled binary
COPY --from=builder /out/helpdesk-server /app/server

# Create upload directory
RUN mkdir -p /app/storage/uploads && \
    chown -R app:app /app

ENV TZ=Asia/Jakarta

EXPOSE 3000

USER app

ENTRYPOINT ["/app/server"]

# Cobra command registered in server.go
CMD ["httpsrv"]