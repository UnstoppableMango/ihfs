{
  perSystem =
    {
      inputs',
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
        modules = ./gomod2nix.toml;
      };
    };
}
