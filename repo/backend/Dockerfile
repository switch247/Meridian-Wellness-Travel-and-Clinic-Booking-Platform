FROM golang:1.23-alpine AS builder
WORKDIR /src
COPY go.mod ./go.mod
COPY backend ./backend
COPY tests ./tests
RUN cd backend && go mod download
RUN cd backend && go build -o /src/server ./cmd/server

FROM golang:1.23-alpine
WORKDIR /app
RUN adduser -D appuser
COPY --from=builder /src/server /app/server
COPY --from=builder /src/backend/migrations /app/migrations
COPY --from=builder /src/backend/seed /app/seed
COPY --from=builder /src/backend/docs /app/docs
COPY --from=builder /src/backend/run_tests.sh /app/run_tests.sh
COPY --from=builder /src/go.mod /app/go.mod
COPY --from=builder /src/tests /app/tests
COPY --from=builder /src/backend /app/backend
RUN chmod +x /app/run_tests.sh
USER appuser
EXPOSE 8080
CMD ["/app/server"]
