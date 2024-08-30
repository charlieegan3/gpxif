{
  description = "gpxif";

  inputs = {
    nixpkgs = {
      type = "github";
      owner = "NixOS";
      repo = "nixpkgs";
      rev = "4cf7951a91440879f61e05460441762d59adc017";
    };
    gomod2nix = {
      type = "github";
      owner = "nix-community";
      repo = "gomod2nix";
      rev = "4e08ca09253ef996bd4c03afa383b23e35fe28a1";
    };
    flake-utils = {
      type = "github";
      owner = "numtide";
      repo = "flake-utils";
      rev = "b1d9ab70662946ef0850d488da1c9019f3a9752a";
    };
  };

  outputs =
    {
      nixpkgs,
      flake-utils,
      gomod2nix,
      ...
    }:
    let
      utils = flake-utils;
    in
    utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = import nixpkgs { inherit system; };
        mkGoEnv = gomod2nix.legacyPackages.${system}.mkGoEnv;
        goEnv = mkGoEnv { pwd = ./.; };
        buildGoApplication = gomod2nix.legacyPackages.${system}.buildGoApplication;

      in
      {
        packages.default = buildGoApplication {
          pname = "gpxif";
          version = "0.1";
          pwd = ./.;
          src = ./.;
          modules = ./gomod2nix.toml;
          checkPhase = ''
            go test -run '^TestCheckModTime' -v ./...
          '';
        };

        devShells = {
          default = pkgs.mkShell {
            packages = with pkgs; [
              go_1_22
              golangci-lint
              gomod2nix.packages.${system}.default
              goEnv
            ];
          };
        };
      }
    );
}
