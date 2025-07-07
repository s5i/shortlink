# shortlink

## Installation

```sh
# Choose a path for local files.
SHORTLINK_PATH="/docker/shortlink"

# Change as desired.
tee compose.yaml << EOF > /dev/null
services:
  shortlink:
    container_name: shortlink
    image: shyym/shortlink:latest
    restart: always
    ports:
      - 3000:3000
    volumes:
      - /docker/shortlink:/appdata
EOF

mkdir -p ${SHORTLINK_PATH}
sudo docker compose up --pull=always --force-recreate --detach

# Edit the config.
${EDITOR:-vi} "${SHORTLINK_PATH}/config.yaml"

# Remove the sentinel to start the service.
rm "${SHORTLINK_PATH}/SENTINEL.readme"


sudo docker compose up --pull=always --force-recreate --detach
```
