# Build context: core/libr
FROM golang:1.24.4-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /librctl ./cmd/librctl

FROM alpine:3.20
COPY --from=build /librctl /librctl
ENTRYPOINT ["/librctl"]
