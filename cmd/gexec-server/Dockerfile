FROM --platform=$BUILDPLATFORM golang:1.26.4-alpine@sha256:7a3e50096189ad57c9f9f865e7e4aa8585ed1585248513dc5cda498e2f41812c AS builder

RUN apk add --no-cache -U nodejs npm git curl
RUN sh -c "$(curl --location https://taskfile.dev/install.sh)" -- -d -b /usr/local/bin

WORKDIR /go/src/gexec
COPY . /go/src/gexec/

RUN --mount=type=cache,target=/go/pkg \
    go mod download -x

ARG TARGETOS
ARG TARGETARCH

ENV ASDF_VERSION=v0.16.0

RUN --mount=type=cache,target=/go/pkg \
    --mount=type=cache,target=/root/.cache/go-build \
    task fe:install fe:generate fe:build && \
    task be:generate build:server GOOS=${TARGETOS} GOARCH=${TARGETARCH} && \
    GOOS=${TARGETOS} GOARCH=${TARGETARCH} go install -ldflags="-s -X main.version=${ASDF_VERSION}" github.com/asdf-vm/asdf/cmd/asdf@${ASDF_VERSION} && \
    mv $(find /go/bin -iname asdf) /go/bin # doing this as multi-arch is using different target paths

FROM alpine:3.24@sha256:a2d49ea686c2adfe3c992e47dc3b5e7fa6e6b5055609400dc2acaeb241c829f4

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

COPY --from=builder --chown=1337:1337 /go/src/gexec/bin/gexec-server /usr/bin/gexec-server
COPY --from=builder --chown=1337:1337 /go/src/gexec/pkg/frontend/files /usr/share/gexec
COPY --from=builder --chown=1337:1337 /go/bin/asdf /usr/bin/asdf
WORKDIR /var/lib/gexec
USER gexec

RUN asdf plugin add terraform https://github.com/asdf-community/asdf-hashicorp.git && \
    asdf plugin add opentofu https://github.com/kc-workspace/asdf-opentofu && \
    asdf plugin add ansible https://github.com/kc-workspace/asdf-ansible.git
