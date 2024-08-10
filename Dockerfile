FROM golang:1.22-alpine as builder
ENV CGO_ENABLED=0
WORKDIR /go/src/
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -ldflags '-w -s' -v -o /usr/local/bin/kustomize-dot ./cmd/kustomize-dot

FROM alpine:latest
COPY --from=builder /usr/local/bin/kustomize-dot /usr/local/bin/kustomize-dot
ENTRYPOINT ["kustomize-dot", "plugin"]
