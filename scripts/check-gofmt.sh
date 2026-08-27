#!/bin/sh
set -eu

unformatted=$(gofmt -l "$@")
if [ -z "$unformatted" ]; then
    exit 0
fi

printf 'Run gofmt on:\n%s\n' "$unformatted"
exit 1
