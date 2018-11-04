#! /bin/bash

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

apps=(
    busybox
    debian
    ubuntu
    archlinux
    fedora
)

for app in ${apps[@]}; do
    echo "******************************************"
    echo "testing ${app}..."
    echo "******************************************"
    echo
    "${SCRIPT_DIR}/../containerflight" run "${SCRIPT_DIR}/${app}.yaml" || exit -1
    echo
done

exit 0
