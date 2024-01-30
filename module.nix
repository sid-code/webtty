webtty: {
  config,
  lib,
  pkgs,
  ...
}: let
  inherit (lib) types mkEnableOption mkOption mkIf;
  cfg = config.services.webtty;
in {
  options.services.webtty = {
    enable = mkEnableOption "WebTTY";
    package = mkOption {
      type = types.package;
      description = "The webtty package to use.";
      default = webtty;
    };
    config = mkOption {
      type = types.submodule {
        options = {
          oneWay = mkOption {
            type = types.bool;
            default = false;
          };
          verbose = mkOption {
            type = types.bool;
            default = true;
          };
          nonInteractive = mkOption {
            type = types.bool;
            default = true;
          };
          stunServer = mkOption {
            type = types.str;
            default = "stun:stun.l.google.com:19302";
          };
          cmd = mkOption {
            type = types.str;
            default = "bash";
          };
          httpPort = mkOption {
            type = types.port;
            default = 3247;
            description = "The port on which webtty should listen.";
          };
        };
      };
    };
  };

  config = mkIf cfg.enable {
    systemd.services.webtty = {
      wants = ["network.target"];
      wantedBy = ["multi-user.target"];
      serviceConfig = {
        ExecStart = pkgs.writeShellScript "launcher" ''
          ${cfg.package}/bin/webtty ${
            (pkgs.formats.toml {}).generate "config.toml" cfg.config
          }
        '';
      };
    };

    networking.firewall.allowedUDPPortRanges = [
      {
        from = 30000;
        to = 50000;
      }
    ];
  };
}
