# linkbot

A Matrix bot built on [ragecore](https://github.com/dknr/ragecore) that replies to
messages containing links with an Open Graph preview.

- Finds the first URL in a message whose host is on the whitelist (`github.com`
  and `x.com`, plus subdomains), fetches the page, extracts `og:title`,
  `og:description` and `og:url`, and replies with a plain-text preview.
- Redirects are followed only while the destination host stays whitelisted.
- Uses ragecore for login, E2EE, device verification, and auto-joining rooms.

## Build

```sh
go build -tags goolm -o linkbot .   # goolm avoids the libolm cgo dependency
```

`ragecore` is referenced via a `replace` directive in `go.mod`; point it at your
checkout if it moves.

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

## Run

```sh
./linkbot -config config.json
```

The whitelist lives in `preview.go` (`allowedDomains`); edit and rebuild to
change it.
