# linkbot

A Matrix bot built on [ragecore](https://github.com/dknr/ragecore) that replies to
messages containing links with an Open Graph preview.

- Finds the first URL in a message whose host is on the whitelist (`github.com`
  and `x.com`, plus subdomains), fetches the page, extracts `og:title`,
  `og:description` and `og:url`, and replies with a styled HTML preview
  (title as link, description in a blockquote).
- Redirects are followed only while the destination host stays whitelisted.
- Uses ragecore for login, E2EE, and device verification.

## Build

```sh
make build   # builds to ./linkbot
```

Or manually: `go build -tags goolm -o linkbot .` (`goolm` avoids the `libolm`
cgo dependency).

## Deploy

### Systemd service

A unit file is included at `contrib/linkbot.service`.

```bash
# Create dedicated service user
sudo useradd --system --no-create-home --shell /usr/sbin/nologin linkbot
sudo mkdir -p /etc/linkbot /var/lib/linkbot
sudo chown linkbot:linkbot /var/lib/linkbot
sudo cp config.json /etc/linkbot/config.json
sudo chown linkbot:linkbot /etc/linkbot/config.json

# Build and install
make build
sudo cp linkbot /usr/local/bin/linkbot

# Enable and start
sudo systemctl daemon-reload
sudo systemctl enable --now linkbot.service

# Monitor
journalctl -u linkbot.service -f
```

The unit expects:
- Binary at `/usr/local/bin/linkbot`
- Config at `/etc/linkbot/config.json`
- Database dir `/var/lib/linkbot` (writable by `linkbot` user)

## Configure

Create `config.json` (or pass `-config <path>`):

```json
{
  "homeserver": "https://matrix.example.org",
  "user": "@bot:example.org",
  "password": "change-me",
  "recovery_key": "",
  "database": "linkbot.db",
  "pickle_key": "linkbot"
}
```

`homeserver`, `user`, `password` are required. `recovery_key` is only needed to
re-verify the bot's device from key backup; if empty and no keys exist, the bot
generates new cross-signing keys and logs the recovery key on startup.

## Local development

```sh
make run   # build and run locally
```

For manual invocation: `./linkbot -config config.json`.

The whitelist lives in `preview.go` (`allowedDomains`); edit and rebuild to
change it.
