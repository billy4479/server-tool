{
  pkgs,
  lib,
  server-tool-unwrapped,
  symlinkJoin,
  enabledJavaVersions ? [
    "8"
    "17"
    "21"
  ],
}:
let
  jresEnv = map (
    version: "--set 'JAVA_${version}' '${lib.getExe pkgs."jdk${version}"}'"
  ) enabledJavaVersions;

  pathPrefix = lib.makeBinPath (
    with pkgs;
    [
      zenity
      git
    ]
  );
in
symlinkJoin rec {
  name = "${pname}-${version}";
  pname = "server-tool";
  inherit (server-tool-unwrapped) version;
  paths = [ server-tool-unwrapped ];

  nativeBuildInputs = [ pkgs.wrapGAppsHook4 ];

  postBuild = ''
    gappsWrapperArgs+=(
      --prefix PATH : ${pathPrefix}
      ${lib.concatStringsSep " \n" jresEnv}
    )

    wrapGAppsHook
  '';

  inherit (server-tool-unwrapped) meta;
}
