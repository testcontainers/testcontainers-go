#!/usr/bin/env bash

BASEDIR=$(dirname "$0")

go_fmt() {
    format=$(go fmt ./...)
    if [[ ${format} ]]; then
        echo "go fmt failed:"
        echo "${format}"
        exit 1
    else
        echo "go fmt passed"
    fi
}

main() {
    cd "${BASEDIR}"/.. || return

    go_fmt
}

main "$@"