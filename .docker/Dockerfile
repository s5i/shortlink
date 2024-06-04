FROM alpine:latest

WORKDIR /app
ADD --chmod=750 https://github.com/s5i/shortlink/releases/latest/download/shortlink /app/shortlink
ADD --chmod=750 https://raw.githubusercontent.com/s5i/shortlink/main/.docker/entrypoint.sh /app/entrypoint.sh

WORKDIR /appdata
VOLUME /appdata
EXPOSE 3000

ENTRYPOINT /app/entrypoint.sh