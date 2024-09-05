{ lib, buildGoModule }:
buildGoModule rec {
  version = "2.1.4";
  pname = "server-tool-unwrapped";

  src = lib.cleanSource ./..;

  ldflags = [
    "-s"
    "-w"
    "-X \"github.com/billy4479/server-tool/lib.Version=${version}\""
  ];

  vendorHash = "sha256-YoguesTG55+Cl5ieCF3FFQK3B6EMpjGNmEV8QHu1VKE=";

  meta = with lib; {
    description = "A tool to manage Minecraft servers";
    homepage = "https://github.com/billy4479/server-tool";
    license = licenses.mit;
  };
}
