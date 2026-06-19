FROM --platform=$BUILDPLATFORM golang:1.26.4-alpine@sha256:3ad57304ad93bbec8548a0437ad9e06a455660655d9af011d58b993f6f615648 AS builder

RUN apk add --no-cache -U git curl
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
    task be:generate build:runner GOOS=${TARGETOS} GOARCH=${TARGETARCH} && \
    GOOS=${TARGETOS} GOARCH=${TARGETARCH} go install -ldflags="-s -X main.version=${ASDF_VERSION}" github.com/asdf-vm/asdf/cmd/asdf@${ASDF_VERSION} && \
    mv $(find /go/bin -iname asdf) /go/bin # doing this as multi-arch is using different target paths

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

COPY --from=builder --chown=1337:1337 /go/src/gexec/bin/gexec-runner /usr/bin/gexec-runner
COPY --from=builder --chown=1337:1337 /go/bin/asdf /usr/bin/asdf
WORKDIR /var/lib/gexec
USER gexec

RUN asdf plugin add terraform https://github.com/asdf-community/asdf-hashicorp.git && \
    asdf plugin add opentofu https://github.com/kc-workspace/asdf-opentofu && \
    asdf plugin add ansible https://github.com/kc-workspace/asdf-ansible.git
