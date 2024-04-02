{
  pkgs,
  lib,
  buildGoModule,
  enabledJavaVersions ? ["8" "17"],
}: let
  pname = "server-tool";
  version = "2.1.4";

  jres =
    map (
      version: "--set 'JAVA_${version}' '${lib.getExe pkgs."jdk${version}"}'"
    )
    enabledJavaVersions;
in
  buildGoModule rec {
    inherit pname version;
    src = lib.cleanSource ./.;
    nativeBuildInputs = [pkgs.makeWrapper];

    ldflags = ["-s" "-w" "-X 'github.com/billy4479/server-tool/lib.Version=${version}'"];

    vendorHash = "sha256-YoguesTG55+Cl5ieCF3FFQK3B6EMpjGNmEV8QHu1VKE=";

    buildPhase = ''
      go build -o ${pname} .
    '';

    patches = [./0001-Force-system-s-java.patch];

    installPhase = ''
      install -Dm755 ${pname} $out/bin/${pname}
    '';

    postFixup =
      ''
        wrapProgram $out/bin/${pname} \
                    --prefix PATH : ${lib.makeBinPath [pkgs.gnome.zenity]} \
      ''
      + lib.concatStringsSep " \\\n" jres;

    meta = with lib; {
      description = "A tool to manage Minecraft servers";
      homepage = "https://github.com/billy4479/server-tool";
      license = licenses.mit;
    };
  }
