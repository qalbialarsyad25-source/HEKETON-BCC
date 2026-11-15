ARG GO_VERSION=1.27

FROM golang:${GO_VERSION}-alpine AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/server .

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
