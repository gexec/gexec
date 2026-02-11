FROM --platform=$BUILDPLATFORM golang:1.26.0-alpine@sha256:d4c4845f5d60c6a974c6000ce58ae079328d03ab7f721a0734277e69905473e5 AS builder

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
