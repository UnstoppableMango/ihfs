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
      packages.mockfs = buildGoApplication {
        pname = "mockfs";
        version = "0.0.1";
        src = lib.cleanSource ./.;
        modules = ./gomod2nix.toml;
        go = pkgs.go_1_26;
        doCheck = false;
      };
    };
}
