# syntax=docker/dockerfile:1
FROM golang:1.24.4-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./
RUN CGO_ENABLED=0 go build -o /relay .

FROM alpine:3.20
COPY --from=build /relay /relay
ENTRYPOINT ["/relay"]
