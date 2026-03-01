#!/usr/bin/env bash
goreleaser release --snapshot --clean --skip=validate 2>&1 | tail -10
