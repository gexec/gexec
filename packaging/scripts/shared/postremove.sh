#!/bin/sh
set -e

if [ ! -d /var/lib/gexec ] && [ ! -d /etc/gexec ]; then
    userdel gexec 2>/dev/null || true
    groupdel gexec 2>/dev/null || true
fi
