{
  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  outputs =
    { self, nixpkgs }:
    let
      system = "x86_64-linux";
      pkgs = import nixpkgs {
        inherit system;
      };
      tree-sitter-nu = pkgs.callPackage ./tree-sitter-nu.nix { };
      tree-sitter-dir = pkgs.callPackage ./tree-sitter-dir.nix {
        inherit tree-sitter-nu;
      };
    in
    {
      devShells.${system}.default =
        let
          libs = with pkgs; [ ];
        in
        pkgs.mkShell {
          name = "devenv";
          buildInputs = libs;
          nativeBuildInputs = (
            with pkgs;
            [
              pkg-config
              tree-sitter
              tree-sitter-grammars.tree-sitter-nu
            ]
          );

          LD_LIBRARY_PATH = "${pkgs.lib.makeLibraryPath libs}:$LD_LIBRARY_PATH";

          shellHook = ''
            export CGO_ENABLED=1
            export TREE_SITTER_DIR="${tree-sitter-dir}"
            echo "Devshell activated."
          '';
        };
      ts-nu = "${pkgs.tree-sitter-grammars.tree-sitter-nu}";
    };
}
