#! /bin/bash

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

apps=(
    busybox
    debian
    ubuntu
    archlinux
    fedora
)

cd "${SCRIPT_DIR}/../" && go build

for app in ${apps[@]}; do
    echo "******************************************"
    echo "testing ${app}..."
    echo "******************************************"
    echo
    "${SCRIPT_DIR}/../containerflight" run "${SCRIPT_DIR}/${app}.yaml" || exit -1
    echo
done

exit 0
