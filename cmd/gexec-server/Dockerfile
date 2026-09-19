FROM --platform=$BUILDPLATFORM debian:13-slim@sha256:d7e12182ce18b85b93007c1dedf31f2d29e01ccf3182cc4017c709b6259bc132 AS builder
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
    task fe:install fe:generate fe:build && \
    task be:generate build:server GOOS=${TARGETOS} GOARCH=${TARGETARCH} && \
    GOBUILD="$(mise where go)/bin/go" && \
    unset GOBIN && CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} "${GOBUILD}" install -tags netgo -ldflags "-s -w -extldflags '-static' -X main.version=${ASDF_VERSION}" github.com/asdf-vm/asdf/cmd/asdf@${ASDF_VERSION} && \
    GOPATHBIN="$("${GOBUILD}" env GOPATH)/bin" && \
    if [ -f "${GOPATHBIN}/${TARGETOS}_${TARGETARCH}/asdf" ]; then \
      mv "${GOPATHBIN}/${TARGETOS}_${TARGETARCH}/asdf" /src/bin/asdf; \
    else \
      mv "${GOPATHBIN}/asdf" /src/bin/asdf; \
    fi && \
    /src/bin/asdf --version

FROM alpine:3.24@sha256:294b683cb724975bec92580e1e685676bd4b50bda910ddb8c51d4cabeaec77e6

RUN apk add --no-cache ca-certificates mailcap python3 python3-dev coreutils openssh-client-default bash git gnupg curl wget tar rsync sshpass tini tzdata zip unzip jq && \
    addgroup -g 1337 gexec && \
    adduser -D -u 1337 -h /var/lib/gexec -G gexec gexec && \
    install -d -o gexec -g gexec /var/lib/asdf

EXPOSE 8000 8080
VOLUME ["/var/lib/gexec"]
ENTRYPOINT ["/usr/bin/gexec-server"]
CMD ["start"]
HEALTHCHECK CMD ["/usr/bin/gexec-server", "health"]

ENV GEXEC_DATABASE_DRIVER="sqlite3"
ENV GEXEC_DATABASE_NAME="/var/lib/gexec/database.sqlite3"
ENV GEXEC_UPLOAD_DRIVER="file"
ENV GEXEC_UPLOAD_PATH="/var/lib/gexec/uploads"

ENV ASDF_DATA_DIR="/var/lib/asdf"
ENV PATH="${ASDF_DATA_DIR}/shims:$PATH"

COPY --from=builder --chown=1337:1337 /src/bin/asdf /usr/bin/asdf
COPY --from=builder --chown=1337:1337 /src/bin/gexec-server /usr/bin/gexec-server
COPY --from=builder --chown=1337:1337 /src/pkg/frontend/files /usr/share/gexec
WORKDIR /var/lib/gexec
USER gexec

RUN asdf plugin add terraform https://github.com/asdf-community/asdf-hashicorp.git && \
    asdf plugin add opentofu https://github.com/kc-workspace/asdf-opentofu.git && \
    asdf plugin add ansible https://github.com/kc-workspace/asdf-ansible.git
