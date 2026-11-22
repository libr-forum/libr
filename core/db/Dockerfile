# syntax=docker/dockerfile:1
FROM golang:1.24.4-alpine

# Install required packages
# go-sqlite3 needs cgo
RUN apk add --no-cache git build-base

# Set working directory for build
WORKDIR /build

# Copy modules
COPY core/db /build/core/db
COPY core/libr /build/core/libr

# Move into db module
WORKDIR /build/core/db

RUN go mod download

# Build the binary to a clean runtime path
RUN CGO_ENABLED=1 go build -o /app/dbnode .

# Use a minimal final image (optional: smaller, safer)
# If you want a separate stage, you can use `golang:alpine` again instead of scratch
FROM alpine:latest

# Install minimal runtime needs
RUN apk add --no-cache libc6-compat

# Copy the binary from build stage
COPY --from=0 /app/dbnode /dbnode

# Set working directory
WORKDIR /

# Ensure binary is executable
RUN chmod +x /dbnode

# Run it
CMD ["/dbnode"]
