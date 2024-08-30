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
    pre-commit-hooks = {
      type = "github";
      owner = "cachix";
      repo = "git-hooks.nix";
      rev = "c7012d0c18567c889b948781bc74a501e92275d1";
    };
  };

  outputs =
    {
      self,
      nixpkgs,
      flake-utils,
      gomod2nix,
      pre-commit-hooks,
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
        checks = {
          pre-commit-check = pre-commit-hooks.lib.${system}.run {
            src = ./.;
            hooks = {
              dprint = {
                enable = true;
                name = "dprint check";
                entry = "dprint check --allow-no-files";
              };
              nixfmt = {
                enable = true;
                name = "nixfmt check";
                entry = "nixfmt -c ";
                types = [ "nix" ];
              };
            };
          };
        };

        packages.default = buildGoApplication {
          pname = "gpxif";
          version = "0.1";
          pwd = ./.;
          src = ./.;
          modules = ./gomod2nix.toml;
          checkPhase = ''
            NIX_BUILD=true go test -v ./...
          '';
        };

        devShells = {
          default = pkgs.mkShell {
            inherit (self.checks.${system}.pre-commit-check) shellHook;
            buildInputs = self.checks.${system}.pre-commit-check.enabledPackages;

            packages = with pkgs; [
              go_1_22
              golangci-lint
              gomod2nix.packages.${system}.default
              goEnv

              dprint
              nixfmt-rfc-style
            ];
          };
        };
      }
    );
}
