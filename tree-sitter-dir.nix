{ tree-sitter-nu, stdenv }:

stdenv.mkDerivation {
  name = "tree-sitter-dir";

  dontUnpack = true;

  installPhase = ''
    mkdir -p $out

    ln -s ${tree-sitter-nu} $out/tree-sitter-nu
    echo "{\"parser-directories\": [\"$out\"]}" > $out/config.json
  '';
}
