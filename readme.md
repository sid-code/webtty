
## WebTTY

WebTTY allows you to share a terminal session from your machine using
WebRTC.

This is a fork of [WebTTY](https://github.com/maxmcd/webtty) that
caters to my use case. Some choices I made:

  - Web-based frontends (to require zero setup on client machine)
  - WebRTC with no TURN server (for minimal latency) 
  - Config file over CLI arguments (no good reason, only because it
    simplifies the code a little)

### Status

This works pretty well (for me).

### Running

Create a `config.toml` that contains the following:

```toml
oneWay = false
verbose = true
nonInteractive = true
stunServer = "stun:stun.l.google.com:19302"
cmd = "bash"
httpPort = 3247
```

Then run:

```shell
webtty config.toml
```

or with the flake output:

```shell
nix run github:sid-code/webtty -- config.toml
```

Then open up `localhost:3247`. Add a reverse proxy and TLS.
