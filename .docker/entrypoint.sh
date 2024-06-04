#!/bin/sh

test  -f "/appdata/shortlink.yaml" || /app/shortlink --create_config
/app/shortlink