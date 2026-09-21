{
  description = "Portable dev shell";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs {
          inherit system;
        };
      in {
        devShells.default = pkgs.mkShell {
          packages = [

          ];
        };

        apps.default = {
          type = "app";
          program = let
            script = pkgs.writeShellScript "run-maven" ''
              export JAVA_HOME=${pkgs.jdk17_headless}
              export PATH=${pkgs.jdk17_headless}/bin:${pkgs.maven}/bin:$PATH
              exec mvn spring-boot:run
            '';
          in "${script}";
        };
      }
    );
}
