{
  perSystem =
    {
      inputs',
      pkgs,
      lib,
      ...
    }:
    let
      inherit (inputs'.gomod2nix.legacyPackages) buildGoApplication;
    in
    {
      packages.ctrfs = buildGoApplication {
        pname = "ctrfs";
        version = "0.0.1";
        src = lib.cleanSource ./.;
        go = pkgs.go_1_26;
        modules = ./gomod2nix.toml;
      };
    };
}
