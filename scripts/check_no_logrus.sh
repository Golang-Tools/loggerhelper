#!/usr/bin/env bash
# 校验 v4 不依赖 logrus(纯 slog 实现)。每次改动后运行一次。
set -euo pipefail
cd "$(dirname "$0")/.."
if go list -deps ./... | grep -q logrus; then
    echo "FAIL: v4 不得依赖 logrus" >&2
    exit 1
fi
echo "PASS: 零 logrus 依赖"
