{
  lib,
  buildGoModule,
  go,
  gnome,
}: let
  pname = "server-tool";
  version = "2.1.4";
in
  buildGoModule rec {
    inherit pname version;
    src = lib.cleanSource ./.;
    nativeBuildInputs = [go];
    buildInputs = [gnome.zenity];

    vendorHash = "sha256-YoguesTG55+Cl5ieCF3FFQK3B6EMpjGNmEV8QHu1VKE=";

    buildPhase = ''
      go build -o ${pname} .
    '';

    installPhase = ''
      install -Dm755 ${pname} $out/bin/${pname}
    '';

    meta = with lib; {
      description = "A tool to manage Minecraft servers";
      homepage = "https://github.com/billy4479/server-tool";
      license = licenses.mit;
    };
  }
