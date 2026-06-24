ARG VERSION=dev

FROM golang:1.22-alpine AS builder
ARG VERSION
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags "-X main.version=${VERSION} -s -w" -o /bin/kube-janitor ./cmd/janitor

FROM gcr.io/distroless/static-debian12:latest
COPY --from=builder /bin/kube-janitor /kube-janitor
USER nobody
ENTRYPOINT ["/kube-janitor"]
