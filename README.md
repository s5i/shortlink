# shortlink

## Installation

```sh
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
sudo docker compose up --pull=always --force-recreate --detach
```
