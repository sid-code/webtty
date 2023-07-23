
## WebTTY

WebTTY allows you to share a terminal session from your machine using
WebRTC.

This is a heavily modified version of
[WebTTY](https://github.com/maxmcd/webtty) that focuses on the
in-browser use case.

### Status

There are a handful of bugs to fix, but everything works pretty well
at the moment. Please open an issue if you find a bug.

### Installation

TODO

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

Then open up `localhost:3247`. Add a reverse proxy and TLS.


### Terminal Size

By default WebTTY forces the size of the client terminal. This means the host size can frequently render incorrectly. One way you can fix this is by using tmux:

```bash
tmux new-session -s shared
# in another terminal
webtty -ni -cmd tmux attach-session -t shared
```
Tmux will now resize the session to the smallest terminal viewport.

