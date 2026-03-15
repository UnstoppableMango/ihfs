{
  description = "I ❤️ FileSystems!";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs?ref=nixos-unstable";
    systems.url = "github:nix-systems/default";

    flake-parts = {
      url = "github:hercules-ci/flake-parts";
      inputs.nixpkgs-lib.follows = "nixpkgs";
    };

    gomod2nix = {
      url = "github:nix-community/gomod2nix";
      inputs.nixpkgs.follows = "nixpkgs";
      inputs.flake-utils.inputs.systems.follows = "systems";
    };

    treefmt-nix = {
      url = "github:numtide/treefmt-nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs =
    inputs@{ flake-parts, ... }:
    flake-parts.lib.mkFlake { inherit inputs; } {
      systems = import inputs.systems;

      imports = [
        inputs.treefmt-nix.flakeModule

        ./ctrfs
        ./ghfs
        ./mockfs
      ];

      perSystem =
        {
          inputs',
          pkgs,
          lib,
          ...
        }:
        let
          inherit (inputs'.gomod2nix.legacyPackages) buildGoApplication gomod2nix mkGoEnv;

          go = pkgs.go_1_26;
          goEnv = mkGoEnv { pwd = ./.; };

          ihfs = buildGoApplication {
            pname = "ihfs";
            version = "0.0.1";

            src = lib.cleanSourceWith {
              src = lib.cleanSource ./.;
              filter =
                path: type:
                !(lib.any (prefix: lib.hasPrefix prefix path) (
                  map toString [
                    ./ctrfs
                    ./ghfs
                    ./mockfs
                  ]
                ));
            };

            go = go;
            modules = ./gomod2nix.toml;
          };
        in
        {
          packages = {
            inherit ihfs;
            default = ihfs;
          };

          devShells.default = pkgs.mkShellNoCC {
            packages = with pkgs; [
              bashInteractive
              ginkgo
              gnumake
              go
              goEnv
              golangci-lint
              gomod2nix
              gopls
              goreleaser
              nixfmt
              ripgrep
              watchexec
            ];

            GINKGO = "${pkgs.ginkgo}/bin/ginkgo";
            GO = "${go}/bin/go";
            GOMOD2NIX = "${gomod2nix}/bin/gomod2nix";
            GOPLS = "${pkgs.gopls}/bin/gopls";
            GORELEASER = "${pkgs.goreleaser}/bin/goreleaser";
            NIXFMT = "${pkgs.nixfmt}/bin/nixfmt";
          };

          treefmt.programs = {
            nixfmt.enable = true;
            gofmt.enable = true;
          };
        };
    };
}
