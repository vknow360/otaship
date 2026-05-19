# Stage 1: Build the Go binary using a Go build environment based on Alpine Linux for a small footprint
FROM golang:1.25-alpine AS builder

# Set the working directory inside the builder container where our code will live
WORKDIR /app

# Copy dependency definition files first (go.mod and go.sum)
# This is done before copying the source code to leverage Docker layer caching
# If dependencies haven't changed, Docker will reuse the cached layer and skip downloading dependencies
COPY go.mod go.sum ./

# Download the Go dependencies specified in go.mod
RUN go mod download

# Copy all the project source code, migrations, and assets from the host to the container's working directory
COPY . .

# Compile the Go application as a statically linked binary (CGO_ENABLED=0) targeting Linux
# Statically compiling the binary allows it to run without dynamic library dependencies in a minimal scratch or alpine image
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server

# Stage 2: Create a minimal, secure runtime image to host our compiled binary
FROM alpine:3.20

# Install CA certificates to enable secure outgoing HTTPS requests (essential for external services like S3 or Cloudinary APIs)
RUN apk add --no-cache ca-certificates

# Set the working directory in the final runtime container
WORKDIR /app

# Copy the compiled executable from the builder stage to our minimal runtime image
COPY --from=builder /app/server .

# Copy the database migrations folder from the builder stage to enable migrations on startup
COPY --from=builder /app/migrations ./migrations

# Copy the OpenAPI specification file so the server can serve API documentation to the user
COPY --from=builder /app/openapi.yaml .

# Inform Docker that the container will listen on port 8080 at runtime
EXPOSE 8080

# Specify the default command to execute when the container starts
ENTRYPOINT ["./server"]

