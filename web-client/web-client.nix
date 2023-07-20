{
  stdenv,
  callPackage,
  nodejs,
  nodePackages,
  go,
  buildGoModule,
  bash,
}: let
  nodeDependencies = (callPackage ./default.nix {}).nodeDependencies;
  buildGoModuleWasm = buildGoModule.override {
    go =
      go
      // {
        GOOS = "js";
        GOARCH = "wasm";
      };
  };
  wasmModule = buildGoModuleWasm {
    name = "web-client-wasm";
    version = "unversioned";
    src = ./.;
    vendorHash = "sha256-bC1Cf9+ntETBuKC+8uO3yz9Yq8RQFDxApeBLvU0DlFY=";
  };
  webClient = stdenv.mkDerivation {
    name = "webtty-client";
    src = ./.;
    buildInputs = [nodejs nodePackages.npm go];
    buildPhase = ''
      ln -s ${nodeDependencies}/lib/node_modules ./node_modules
      export PATH="${nodeDependencies}/bin:$PATH"

      sed -i "s:/usr/bin/env.\+:${nodejs}/bin/node:g" ./node_modules/.bin/vite

      # Build the distribution bundle in "dist"
      vite build ./src
    '';

    installPhase = ''
      cp -r dist $out/
      cp ${wasmModule}/bin/js_wasm/web-client $out/main.wasm
    '';
  };
in
  webClient
