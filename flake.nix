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
      imports = [ inputs.treefmt-nix.flakeModule ];

      perSystem =
        { inputs', pkgs, ... }:
        let
          inherit (inputs'.gomod2nix.legacyPackages) buildGoApplication gomod2nix mkGoEnv;

          goEnv = mkGoEnv { pwd = ./.; };
        in
        {
          packages.default = buildGoApplication {
            pname = "ihfs";
            version = "0.0.1";
            src = ./.;
            modules = ./gomod2nix.toml;
          };

          devShells.default = pkgs.mkShellNoCC {
            packages = with pkgs; [
              ginkgo
              go
              goEnv
              gomod2nix
              nixfmt
            ];

            GINKGO = "${pkgs.ginkgo}/bin/ginkgo";
            GO = "${pkgs.go}/bin/go";
            GOMOD2NIX = "${gomod2nix}/bin/gomod2nix";
            NIXFMT = "${pkgs.nixfmt}/bin/nixfmt";
          };

          treefmt.programs = {
            nixfmt.enable = true;
            gofmt.enable = true;
          };
        };
    };
}
