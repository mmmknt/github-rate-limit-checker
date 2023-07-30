# Build stage
FROM golang:1.20 as builder

WORKDIR /app

# Download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Build the application
COPY . ./
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Runtime stage
FROM gcr.io/distroless/static-debian11:latest

WORKDIR /

# Copy the pre-built binary from the build stage
COPY --from=builder /app/main /main

CMD ["/main"]
