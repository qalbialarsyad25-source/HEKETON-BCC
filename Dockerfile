ARG GO_VERSION=1.27

FROM golang:${GO_VERSION}-alpine AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
# Change "." at the end to your main package folder (e.g. ./cmd/server or ./cmd/api)
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/main

FROM alpine:3
RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S app \
    && adduser -S -G app app \
    && mkdir -p /app/logs \
    && chown -R app:app /app

WORKDIR /app
COPY --from=build /out/server /app/server

USER app
EXPOSE 8080
ENTRYPOINT ["/app/server"]