{
  inputs.nixpkgs.url = "github:nixos/nixpkgs/nixpkgs-unstable";
  outputs = {
    self,
    nixpkgs,
    ...
  }: let
    supportedSystems = ["x86_64-linux"];
    forAllSystems = f:
      nixpkgs.lib.genAttrs supportedSystems (system:
        f system (import nixpkgs {inherit system;}));
  in {
    packages = forAllSystems (
      system: pkgs: let
        webClient = pkgs.callPackage ./web-client/web-client.nix {};
      in rec {
        inherit webClient;
        webtty = pkgs.buildGoModule {
          name = "webtty";
          version = "flake-latest";
          src = ./.;
          ldflags = "-X 'main.ServePath=${webClient}'";
          # by default, buildGoModule does a lot of magic
          # that isn't compatible with this repo
          # specifically, the thing where it tries to build all
          # "submodules". For some reason, `go build ./web-client`
          # doesn't work.
          buildPhase = ''
            go build -p "$NIX_BUILD_CORES" --ldflags="$ldflags"
          '';
          installPhase = ''
            mkdir -p $out/bin
            mv ./webtty $out/bin
          '';
          # Don't run any tests
          checkPhase = "";
          vendorHash = "sha256-31Qg8ond4wqo1qokk4Adlv1tuqwNoWYm61Y24DRq48A=";
        };

        default = webtty;
      }
    );

    nixosModules.default =
      import ./module.nix
      self.packages.x86_64-linux.default;

    devShells = forAllSystems (
      system: pkgs: {
        default = pkgs.mkShell {
          nativeBuildInputs = with pkgs; [
            go
            gopls
            nodejs
            nodePackages.npm
            nodePackages.prettier
            emacsPackages.go-mode
            watchexec
          ];
        };
      }
    );
  };
}
