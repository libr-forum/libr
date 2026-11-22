# syntax=docker/dockerfile:1
# Build context: repository root (the relay uses core/libr).
FROM golang:1.24.4-alpine AS build
WORKDIR /src
COPY core/libr core/libr
COPY services/relay services/relay
WORKDIR /src/services/relay
RUN go mod download && CGO_ENABLED=0 go build -o /relay .

FROM alpine:3.20
COPY --from=build /relay /relay
ENTRYPOINT ["/relay"]
