{
  perSystem =
    { inputs', pkgs, lib, ... }:
    let
      inherit (inputs'.gomod2nix.legacyPackages) buildGoApplication gomod2nix;
    in
    {
      packages.ghfs = buildGoApplication {
        pname = "ghfs";
        version = "0.0.1";
        src = lib.cleanSource ./.;
        modules = ./gomod2nix.toml;
      };
    };
}
