# Stage 1: Build
FROM golang:1.24 as builder

WORKDIR /app
COPY . .

RUN go mod download

RUN CGO_ENABLED=0 go build -o manager cmd/main.go

# Stage 2: Minimal runtime image
FROM gcr.io/distroless/static

WORKDIR /
COPY --from=builder /app/manager ./
ENTRYPOINT ["/manager"]
