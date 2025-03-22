#!/bin/sh
set -e

systemctl stop gexec-server.service || true
systemctl disable gexec-server.service || true
