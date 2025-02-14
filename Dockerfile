# Start from the official Golang image as the builder stage
#FROM golang:1.22.4 AS builder
FROM golang:1.22.4-alpine AS builder

# Set the working directory
WORKDIR /app

# Copy the Go module files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the application source code
COPY . .

# Build
RUN GOOS=linux GOARCH=amd64 go build -o main ./cmd

# Use a minimal base image for the final container
FROM alpine:latest

# Set the working directory in the final image
WORKDIR /root/

# Copy the compiled Go binary from the builder stage
COPY --from=builder /app/main .

# Copy the views
COPY views/ views/
COPY static/ static/

# Set the port the container should expose
EXPOSE 8080

# Command to run the application
CMD ["./main"]

