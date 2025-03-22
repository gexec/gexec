#!/bin/sh
set -e

systemctl stop gexec-runner.service || true
systemctl disable gexec-runner.service || true
