#!/bin/sh
set -e

chown -R gexec:gexec /etc/gexec
chown -R gexec:gexec /var/lib/gexec
chmod 750 /var/lib/gexec

if [ -d /run/systemd/system ]; then
    systemctl daemon-reload

    if systemctl is-enabled --quiet gexec-server.service; then
        systemctl restart gexec-server.service
    fi
fi
