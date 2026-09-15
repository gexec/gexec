FROM --platform=$BUILDPLATFORM debian:13-slim AS builder
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

ENV ASDF_VERSION=v0.16.0

RUN --mount=type=cache,target=/root/go/pkg \
    --mount=type=cache,target=/root/.cache/go-build \
    task be:generate build:runner GOOS=${TARGETOS} GOARCH=${TARGETARCH} && \
    GOBUILD="$(mise where go)/bin/go" && \
    unset GOBIN && CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} "${GOBUILD}" install -tags netgo -ldflags "-s -w -extldflags '-static' -X main.version=${ASDF_VERSION}" github.com/asdf-vm/asdf/cmd/asdf@${ASDF_VERSION} && \
    GOPATHBIN="$("${GOBUILD}" env GOPATH)/bin" && \
    if [ -f "${GOPATHBIN}/${TARGETOS}_${TARGETARCH}/asdf" ]; then \
      mv "${GOPATHBIN}/${TARGETOS}_${TARGETARCH}/asdf" /src/bin/asdf; \
    else \
      mv "${GOPATHBIN}/asdf" /src/bin/asdf; \
    fi && \
    /src/bin/asdf --version

FROM alpine:3.24@sha256:28bd5fe8b56d1bd048e5babf5b10710ebe0bae67db86916198a6eec434943f8b

RUN apk add --no-cache ca-certificates mailcap python3 python3-dev coreutils openssh-client-default bash git gnupg curl wget tar rsync sshpass tini tzdata zip unzip jq && \
    addgroup -g 1337 gexec && \
    adduser -D -u 1337 -h /var/lib/gexec -G gexec gexec && \
    install -d -o gexec -g gexec /var/lib/asdf

EXPOSE 8000 8080
VOLUME ["/var/lib/gexec"]
ENTRYPOINT ["/usr/bin/gexec-runner"]
CMD ["start"]

ENV ASDF_DATA_DIR="/var/lib/asdf"
ENV PATH="${ASDF_DATA_DIR}/shims:$PATH"

COPY --from=builder --chown=1337:1337 /src/bin/asdf /usr/bin/asdf
COPY --from=builder --chown=1337:1337 /src/bin/gexec-runner /usr/bin/gexec-runner
WORKDIR /var/lib/gexec
USER gexec

RUN asdf plugin add terraform https://github.com/asdf-community/asdf-hashicorp.git && \
    asdf plugin add opentofu https://github.com/kc-workspace/asdf-opentofu.git && \
    asdf plugin add ansible https://github.com/kc-workspace/asdf-ansible.git
