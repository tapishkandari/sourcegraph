FROM golang:1.15-alpine@sha256:ebac6beacbc2dbae3f454e08745d1d7059e1f061b7c1dca15a4e167224993d99 AS builder

WORKDIR /go/src/resources-report

COPY go.* ./
RUN go mod download

COPY *.go ./
RUN go build -o /bin/resources-report

FROM sourcegraph/alpine:3.12@sha256:133a0a767b836cf86a011101995641cf1b5cbefb3dd212d78d7be145adde636d
# TODO(security): This container should not be running as root!
# hadolint ignore=DL3002
USER root

# hadolint ignore=DL3018
RUN apk add --no-cache ca-certificates

COPY --from=builder /bin/resources-report /usr/local/bin/

ENTRYPOINT ["resources-report"]
