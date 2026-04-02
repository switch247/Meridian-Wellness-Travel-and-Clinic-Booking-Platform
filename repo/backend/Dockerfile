FROM golang:1.23-alpine AS builder
WORKDIR /app/backend
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
RUN go build -o /meridian-server ./cmd/server

FROM golang:1.23-alpine
WORKDIR /app/backend
RUN apk add --no-cache postgresql-client curl
COPY --from=builder /meridian-server /app/meridian-server
COPY backend/ .
COPY backend/migrations /app/migrations

EXPOSE 8080
CMD ["/app/meridian-server"]
