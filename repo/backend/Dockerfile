
FROM golang:1.23-alpine AS builder
WORKDIR /src

# copy everything to simplify build context and leverage layer caching minimally
COPY . .

# ensure modules are downloaded and tidy
RUN cd backend && go mod download && go mod tidy

# build static binary for the server
RUN cd backend && CGO_ENABLED=0 GOOS=linux go build -ldflags "-s -w" -o /src/server ./cmd/server

FROM alpine:3.18
WORKDIR /app

RUN adduser -D appuser

# copy runtime artifacts
COPY --from=builder /src/server /app/server
COPY --from=builder /src/backend/migrations /app/migrations
COPY --from=builder /src/backend/seed /app/seed
COPY --from=builder /src/backend/docs /app/docs
COPY --from=builder /src/backend /app/backend

USER appuser
EXPOSE 8443
CMD ["/app/server"]
