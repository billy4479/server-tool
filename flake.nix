{
  inputs = {
    nixpkgs.url = "nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs =
    {
      self,
      nixpkgs,
      flake-utils,
    }:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
      in
      {
        devShells.default = pkgs.mkShell {
          nativeBuildInputs = with pkgs; [
            go
          ];
          buildInputs = with pkgs; [
            zenity
          ];
        };

        packages = rec {
          default = self.packages.${system}.server-tool;
          server-tool = pkgs.callPackage ./nix { inherit server-tool-unwrapped; };
          server-tool-unwrapped = pkgs.callPackage ./nix/unwrapped.nix { };
        };
      }
    );
}
