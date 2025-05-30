FROM --platform=$BUILDPLATFORM golang:1.24.3-alpine3.21 AS builder

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

FROM alpine:3.22

RUN apk add --no-cache ca-certificates mailcap && \
    addgroup -g 1337 gexec && \
    adduser -D -u 1337 -h /var/lib/gexec -G gexec gexec

VOLUME ["/var/lib/gexec"]
ENTRYPOINT ["/usr/bin/gexec-client"]
CMD ["help"]

COPY --from=builder --chown=1337:1337 /go/src/gexec/bin/gexec-client /usr/bin/gexec-client
WORKDIR /var/lib/gexec
USER gexec
