{
  description = "idefix-go";

  inputs = {
      nixpkgs.url = "nixpkgs/nixos-25.05";
      flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, flake-utils, nixpkgs }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};

        repo = pkgs.fetchFromGitHub {
            owner = "nayarsystems";
            repo = "idefix-go";
            rev = "ca6ed39b84d2fabd5c207291f007f66a3cd99f43";
            hash = "sha256-5//xLu7bx3IXPv/+C867VgfgsPn0YukvS/h3rZO7tKY=";
          };
      in
      rec {
        packages.default = pkgs.buildGoModule rec {
          pname = "idefix";
          src = ./tools/idefix;
          proxyVendor = true;

          ldflags = [
              "-X main.buildFlagVersion=${version}"
          ];

          version = "ca6ed39b84d2fabd5c207291f007f66a3cd99f43";
          vendorHash = "sha256-jSw2hMo2lIgqFD/Dq2eXZ/DxmwRZ30FXqrq9g4PYq0g=";
        };
      });
}
