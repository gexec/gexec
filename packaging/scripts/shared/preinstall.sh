#!/bin/sh
set -e

if ! getent group gexec >/dev/null 2>&1; then
    groupadd --system gexec
fi

if ! getent passwd gexec >/dev/null 2>&1; then
    useradd --system --create-home --home-dir /var/lib/gexec --shell /bin/bash -g gexec gexec
fi
