#!/bin/bash

BINARY="/app/tscanner.app"
EXAMPLE_CONFIG="/app/example_config.yaml"
CONFIG="/appdata/config.yaml"
DB="/appdata/shortlink.db"
SENTINEL="/appdata/SENTINEL.readme"

if [[ ! -f ${CONFIG} ]]; then
    echo "Remove this file after making necessary changes to ${CONFIG}" > "${SENTINEL}"
    cp "${EXAMPLE_CONFIG}" "${CONFIG}"
fi

if [[ -f ${SENTINEL} ]]; then
    echo "Remove ${SENTINEL} to start the service."
fi
until [[ ! -f ${SENTINEL} ]]; do
    sleep 5
done

echo "Starting the service..."
$PWD/app/shortlink.app --config_path "${CONFIG}"
