{
  inputs = {
    self.submodules = true;
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  };
  outputs =
    { self, nixpkgs }:
    let
      system = "x86_64-linux";
      pkgs = import nixpkgs { inherit system; };
      tree-sitter-src = pkgs.fetchFromGitHub {
        owner = "tree-sitter";
        repo = "tree-sitter";
        rev = "16aaed78ae6582ea55a94419828922c7b0960e10";
        hash = "sha256-CsDBpXgAOEzp+i7InNExvK4Syc+7EVabna88Q+N+Lkw=";
      };
    in
    {
      packages.${system}.default = pkgs.buildGoModule {
        pname = "nu-type-alias";
        version = "0.2.4";

        src = ./.;

        vendorHash = "sha256-vqvsMkB0T41XZ0/lj7MbNXUdL4612ThPbqq5rSkjSrM=";
        # sourceRoot = "${./.}/cmd/nu-type-alias";
        subPackages = [ "cmd/nu-type-alias" ];
        meta = {
          mainProgram = "nu-type-alias";
        };

        preBuild = ''
          export CGO_CFLAGS="-I${tree-sitter-src}/lib/include -I${tree-sitter-src}/lib/src $CGO_CFLAGS"
          export CGO_LDFLAGS="$CGO_LDFLAGS"
        '';
      };

      apps.${system}.default = {
        type = "app";
        program = "${self.packages.${system}.default}/bin/nu-type-alias";
      };

      devShells.${system}.default = pkgs.mkShell {
        shellHook = ''
          export CGO_ENABLED=1
        '';
      };
    };
}
