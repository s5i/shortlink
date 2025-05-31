#!/bin/bash

CONFIG_PATH="${PWD}/appdata/shortlink.yaml"
DB_PATH="${PWD}/appdata/shortlink.db"
INIT_PATH="${PWD}/appdata/REMOVE_AFTER_CONFIGURATION"

if [[ ! -f ${CONFIG_PATH} ]]; then
    touch "${INIT_PATH}"
    $PWD/app/shortlink.app --config_path "${CONFIG_PATH}" --create_config
fi

if [[ -f ${INIT_PATH} ]]; then
    echo "Remove ${INIT_PATH} to start the service."
fi
until [[ ! -f ${INIT_PATH} ]]; do
    sleep 5
done

echo "Starting the service..."
$PWD/app/shortlink.app --config_path "${CONFIG_PATH}" --db_path "${DB_PATH}"
