{
  inputs.nixpkgs.url = "github:nixos/nixpkgs/nixpkgs-unstable";
  outputs = {nixpkgs, ...}: let
    supportedSystems = ["x86_64-linux"];
    forAllSystems = f:
      nixpkgs.lib.genAttrs supportedSystems (system:
        f system (import nixpkgs {inherit system;}));
  in {
    devShells = forAllSystems (
      system: pkgs: {
        default = pkgs.mkShell {
          nativeBuildInputs = with pkgs; [
            go
            nodejs
            nodePackages.npm
          ];
        };
      }
    );
  };
}
