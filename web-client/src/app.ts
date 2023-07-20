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

async function waitForDecode(key: string) {
  if (typeof decode !== "undefined") {
    startSession(key);
  } else {
    setTimeout(waitForDecode, 250);
  }
}

window.setTimeout(() => {
  console.log(encode);
  console.log(Go);
}, 1000);

const go = new Go();
WebAssembly.instantiateStreaming(fetch("main.wasm"), go.importObject).then(
  (result) => {
    let inst = result.instance;
    go.run(inst);
  },
);

const startSession = (data: string) => {
  decode(data, (Sdp: string, err: string) => {
    if (err != "") {
      console.log(err);
    }
    console.log("SDP k", data);
    pc.setRemoteDescription(
      new RTCSessionDescription({
        type: "offer",
        sdp: Sdp,
      }),
    ).catch(log);
    pc.createAnswer()
      .then((d) => pc.setLocalDescription(d))
      .catch(log);
  });
};

const term = new Terminal();
term.open(document.getElementById("terminal"));
term.toggleFullScreen();
term.fit();
window.onresize = () => {
  term.fit();
};
term.write("Welcome to the WebTTY web client.\n\r");

let pc = new RTCPeerConnection({
  iceServers: [
    {
      urls: "stun:stun.l.google.com:19302",
    },
  ],
});

let log = (msg: string) => {
  term.write(msg + "\n\r");
};

pc.onnegotiationneeded = (e) => console.log(e);
try {
  (async () => {
    const key = await (await fetch("init", { method: "POST" })).text();
    waitForDecode(key);

    let sendChannel = pc.createDataChannel("data");
    sendChannel.onclose = () => console.log("sendChannel has closed");
    sendChannel.onopen = () => {
      term.reset();
      term.terminadoAttach(sendChannel);
      sendChannel.send(JSON.stringify(["set_size", term.rows, term.cols]));
      console.log("sendChannel has opened");
    };
    // sendChannel.onmessage = e => {}

    pc.onsignalingstatechange = (_) => log("SIGNAL " + pc.signalingState);
    pc.oniceconnectionstatechange = (_) => log("ICE " + pc.iceConnectionState);
    pc.onicecandidate = (event) => {
      if (event.candidate === null) {
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
  })();
} catch (err) {
  console.log(err);
}
