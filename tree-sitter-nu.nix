{
  stdenv,

  tree-sitter,
  nodejs,

  fetchFromGitHub,
}:

stdenv.mkDerivation {
  name = "tree-sitter-nu";

  src = fetchFromGitHub {
    owner = "nushell";
    repo = "tree-sitter-nu";
    rev = "348b787d8b0409091d85fe9d4eb007fe9f3406bb";
    hash = "sha256-OL3fqHjimJ9VrR2UoeIdLxKKcsA1J80A9T8GSBO9KwE=";
  };

  nativeBuildInputs = [
    tree-sitter
    nodejs
  ];

  buildPhase = ''
    tree-sitter generate
  '';
  installPhase = ''
    cp -r $src $out
  '';
}
