{
  inputs.nixpkgs.url = "github:nixos/nixpkgs/nixpkgs-unstable";
  outputs = {nixpkgs, ...}: let
    supportedSystems = ["x86_64-linux"];
    forAllSystems = f:
      nixpkgs.lib.genAttrs supportedSystems (system:
        f system (import nixpkgs {inherit system;}));
  in {
    packages = forAllSystems (
      system: pkgs: rec {
        webtty = pkgs.buildGoModule {
          name = "webtty";
          version = "flake-latest";
          src = ./.;
          vendorHash = "sha256-qhQ+n54AHrD7b4gac6yV7d7A6SqTeyOYQhcENlje/gY=";
        };

        default = webtty;
      }
    );

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
