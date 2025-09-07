# Dockerfile

# Stage 1: Build the Go binary in a temporary container
FROM golang:1.22-alpine AS builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source code
COPY . .

# Build the Go app
# - CGO_ENABLED=0: Disables Cgo to create a statically linked binary.
# -o /app/main: Specifies the output file name.
RUN CGO_ENABLED=0 GOOS=linux go build -v -o /app/main .


# Stage 2: Create the final, lightweight image
FROM alpine:latest

# It's good practice to run as a non-root user for security
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

WORKDIR /home/appuser

# Copy the pre-built binary from the "builder" stage
COPY --from=builder /app/main .

# Expose port 8081 to the outside world
EXPOSE 8081

# Command to run the executable
CMD ["./main"]
