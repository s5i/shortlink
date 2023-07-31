# shortlink

## Installation

```sh
wget https://github.com/s5i/shortlink/releases/latest/download/shortlink
chmod +x ./shortlink

sudo mkdir -p /usr/local/shortlink
sudo mv ./shortlink /usr/local/shortlink
sudo chown root:root /usr/local/shortlink/shortlink

sudo tee /etc/systemd/system/shortlink.service << EOF > /dev/null
[Unit]
Description=shortlink Service
Requires=network.target

[Service]
Type=simple
WorkingDirectory=/usr/local/shortlink 
ExecStart=/usr/local/shortlink/shortlink
Restart=always

[Install]
WantedBy=multi-user.target
EOF

cd /usr/local/shortlink
sudo ./shortlink --create_config

# Change as desired.
sudo vim ./shortlink.yaml

sudo systemctl enable shortlink
sudo systemctl start shortlink
```