{
  stdenv,
  callPackage,
  nodejs,
  nodePackages,
  go,
  buildGoModule,
}: let
  nodeDependencies = (callPackage ./default.nix {}).nodeDependencies;
  wasmModule = buildGoModule {
    name = "web-client-wasm";
    version = "unversioned";
    src = ./src;
    vendorHash = "";
  };
  web-client = stdenv.mkDerivation {
    name = "webtty-client";
    src = ./.;
    buildInputs = [nodejs nodePackages.npm go];
    buildPhase = ''
      ln -s ${nodeDependencies}/lib/node_modules ./node_modules
      export PATH="${nodeDependencies}/bin:$PATH"

      # Build the distribution bundle in "dist"
      npm run build
      cp -r dist $out/
    '';
  };
in
  wasmModule
