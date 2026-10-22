FROM --platform=$BUILDPLATFORM debian:13-slim@sha256:62bd423aeee912639ca8ee681c2b762cb3e25137e0a92b20be455898a9f034ad AS builder
WORKDIR /src

RUN apt-get update && \
    apt-get -y --no-install-recommends install ca-certificates curl git build-essential && \
    rm -rf /var/lib/apt/lists/*

SHELL ["/bin/bash", "-o", "pipefail", "-c"]
ENV PATH="/root/.local/bin:/mise/shims:$PATH" \
    MISE_DATA_DIR="/mise" \
    MISE_CONFIG_DIR="/mise" \
    MISE_CACHE_DIR="/mise/cache" \
    MISE_INSTALL_PATH="/usr/local/bin/mise"

RUN curl --proto '=https' --proto-redir '=https' \
    --fail --show-error --silent --location https://mise.run | sh

COPY mise.toml ./
RUN --mount=type=cache,target=/mise/cache \
   mise trust && mise install

COPY . /src/
RUN --mount=type=cache,target=/root/go/pkg \
    go mod download -x

ARG TARGETOS
ARG TARGETARCH

RUN --mount=type=cache,target=/root/go/pkg \
    --mount=type=cache,target=/root/.cache/go-build \
    task be:generate build:client GOOS=${TARGETOS} GOARCH=${TARGETARCH}

FROM alpine:3.24@sha256:c6b073f8e9e58d881f6ffba6ebf3b9ebd60b9e7d6e8337018c0e68bb5c356873

RUN apk add --no-cache ca-certificates mailcap && \
    addgroup -g 1337 gexec && \
    adduser -D -u 1337 -h /var/lib/gexec -G gexec gexec

VOLUME ["/var/lib/gexec"]
ENTRYPOINT ["/usr/bin/gexec-client"]
CMD ["help"]

COPY --from=builder --chown=1337:1337 /src/bin/gexec-client /usr/bin/gexec-client
WORKDIR /var/lib/gexec
USER gexec
