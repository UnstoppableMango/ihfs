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
      packages.ghfs = buildGoApplication {
        pname = "ghfs";
        version = "0.0.1";
        src = lib.cleanSource ./.;
        go = pkgs.go_1_26;
        modules = ./gomod2nix.toml;
      };
    };
}
