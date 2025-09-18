FROM golang:1.24.5-alpine as BUILDER
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY pkg ./pkg
COPY apps/gateway ./apps/gateway

WORKDIR /src/apps/gateway
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/app ./cmd

FROM alpine
WORKDIR /app
COPY --from=builder /bin/app /app
CMD ["./app"]