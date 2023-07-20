import { Terminal } from "xterm";
import * as attach from "./attach";
import * as fullscreen from "xterm/src/addons/fullscreen/fullscreen";
import * as fit from "xterm/src/addons/fit/fit";

import "xterm/dist/xterm.css";
import "xterm/dist/addons/fullscreen/fullscreen.css";

declare class Go {
  importObject: WebAssembly.Imports;
  run(inst: WebAssembly.Instance): void;
}
declare function decode(
  val: string,
  cb: (res: string, err: string) => void,
): void;
declare function encode(
  val: string,
  cb: (res: string, err: string) => void,
): void;

Terminal.applyAddon(attach);
Terminal.applyAddon(fullscreen);
Terminal.applyAddon(fit);

// Polyfill for WebAssembly on Safari
if (!WebAssembly.instantiateStreaming) {
  WebAssembly.instantiateStreaming = async (resp, importObject) => {
    const source = await (await resp).arrayBuffer();
    return await WebAssembly.instantiate(source, importObject);
  };
}

function poll<T>(
  condition: () => boolean,
  value: () => T,
  pollIntervalMs = 100,
  timeoutMs = 60000,
): Promise<T> {
  return new Promise<T>((resolve, reject) => {
    var interval: any; // TODO: what is this?
    var elapsed = 0;
    const check = () => {
      if (condition()) {
        window.clearInterval(interval);
        resolve(value());
        return;
      }

      elapsed += pollIntervalMs;
      if (elapsed > timeoutMs) {
        reject();
      }
    };

    interval = window.setInterval(check, pollIntervalMs);
  });
}

const go = new Go();
WebAssembly.instantiateStreaming(fetch("main.wasm"), go.importObject).then(
  (result) => {
    let inst = result.instance;
    go.run(inst);
  },
);

async function startSession(pc: RTCPeerConnection, data: string) {
  decode(data, async (Sdp: string, err: string) => {
    if (err != "") {
      console.log(err);
      return;
    }
    await pc
      .setRemoteDescription(
        new RTCSessionDescription({
          type: "offer",
          sdp: Sdp,
        }),
      )
      .catch(log);
    await pc
      .createAnswer()
      .then((d) => pc.setLocalDescription(d))
      .catch(log);
  });
}

const term = new Terminal();
term.open(document.getElementById("terminal"));
term.toggleFullScreen();
term.fit();
window.onresize = () => {
  term.fit();
};
term.write("Welcome to the WebTTY web client.\n\r");

function makePeerConnection(key: string): RTCPeerConnection {
  const pc = new RTCPeerConnection({
    iceServers: [
      {
        urls: "stun:stun.l.google.com:19302",
      },
    ],
  });

  pc.onsignalingstatechange = (_) => log("SIGNAL " + pc.signalingState);
  pc.oniceconnectionstatechange = (_) => log("ICE " + pc.iceConnectionState);
  pc.onicecandidate = (event) => {
    if (event.candidate === null) {
      console.log("ICE CANDIDADO");
      encode(
        pc.localDescription?.sdp ?? "lol",
        (encoded: string, err: string) => {
          if (err != "") {
            console.log(err);
            return;
          }
          fetch(`conn?key=${key}`, { method: "POST", body: encoded }).catch(
            log,
          );
        },
      );
    }
  };
  pc.onnegotiationneeded = (e) => console.log(e);

  return pc;
}

let log = (msg: string) => {
  term.write(msg + "\n\r");
};

try {
  (async () => {
    const key = (await (await fetch("init", { method: "POST" })).text()).trim();
    await poll(
      () => typeof decode !== "undefined",
      async () => {
        const pc = makePeerConnection(key);
        const sendChannel = pc.createDataChannel("data");
        sendChannel.onclose = () => console.log("sendChannel has closed");
        sendChannel.onopen = () => {
          term.reset();
          term.terminadoAttach(sendChannel);
          sendChannel.send(JSON.stringify(["set_size", term.rows, term.cols]));
          console.log("sendChannel has opened");
        };

        await startSession(pc, key);
      },
    );
  })();
} catch (err) {
  console.log(err);
}
