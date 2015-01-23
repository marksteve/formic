# [Formic](https://formic.marksteve.com)

Open-source forms web service written in Go

## Requirements

- A working [Go](http://golang.org/doc/install) environment
- [Godep](https://github.com/tools/godep)
- [Bower](http://bower.io)
- [Redis](http://redis.io)
- [Google OAuth 2.0 Client ID](https://console.developers.google.com/project) (used for login)

## Install

```bash
git clone https://github.com/marksteve/formic
cd formic
bower install
```

## Configuration

Formic uses [drone/config](https://github.com/drone/config) so you can configure it either by creating a `formic.toml` file:

```toml
redis-host = "localhost"
session-secret = "secret"

[google]
client-id = "client id"
client-secret = "client secret"
allowed-emails = "you@company.com,you@gmail.com"
```

Or through environment variables prefixed with `FORMIC_`:

```bash
export FORMIC_REDIS_HOST="localhost"
export FORMIC_SESSION_SECRET="secret"
export FORMIC_GOOGLE_CLIENT_ID="client id"
export FORMIC_GOOGLE_CLIENT_SECRET="client secret"
export FORMIC_GOOGLE_ALLOWED_EMAILS="you@company.com,you@gmail.com"
```

### Google OAuth 2.0

Set your Google OAuth 2.0 Client ID's redirect URI to `http://<ADDRESS>/oauth2callback`.

You can set `google-allowed-emails` to `"anyone"` and host forms for, well, anyone!

## Running

```bash
godep go run main.go
```

Formic binds on `:8000` by default. You can change that using the `-bind` argument:

```bash
godep go run main.go -bind 127.0.0.1:5000
```

## License

[MIT](http://marksteve.mit-license.org)

Space background image from: https://flic.kr/p/9pGfXt

## (Might be) FAQ

__I think I've seen something similar before__

Yes you have!

- https://formkeep.com
- http://forms.brace.io/

__NIH?__

Nope! I was learning Go and wanted to create something that I'd use with it.

__Why name it after the [buggers](http://ansible.wikia.com/wiki/Formic)?__

A friend suggested it. It sounds cool and has "form" in it. It's way better than its original name `go-submit`.
