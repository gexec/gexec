FROM --platform=$BUILDPLATFORM golang:1.25.7-alpine@sha256:f6751d823c26342f9506c03797d2527668d095b0a15f1862cddb4d927a7a4ced AS builder

RUN apk add --no-cache -U git curl
RUN sh -c "$(curl --location https://taskfile.dev/install.sh)" -- -d -b /usr/local/bin

WORKDIR /go/src/gexec
COPY . /go/src/gexec/

RUN --mount=type=cache,target=/go/pkg \
    go mod download -x

ARG TARGETOS
ARG TARGETARCH

RUN --mount=type=cache,target=/go/pkg \
    --mount=type=cache,target=/root/.cache/go-build \
    task be:generate build:client GOOS=${TARGETOS} GOARCH=${TARGETARCH}

FROM alpine:3.23@sha256:25109184c71bdad752c8312a8623239686a9a2071e8825f20acb8f2198c3f659

RUN apk add --no-cache ca-certificates mailcap && \
    addgroup -g 1337 gexec && \
    adduser -D -u 1337 -h /var/lib/gexec -G gexec gexec

VOLUME ["/var/lib/gexec"]
ENTRYPOINT ["/usr/bin/gexec-client"]
CMD ["help"]

COPY --from=builder --chown=1337:1337 /go/src/gexec/bin/gexec-client /usr/bin/gexec-client
WORKDIR /var/lib/gexec
USER gexec
