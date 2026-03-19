{
  description = "A lightweight Subsonic TUI music player built in Go";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";

  outputs = { self, nixpkgs }:
    let
      inherit (nixpkgs) lib;
      forAllSystems = lib.genAttrs lib.systems.flakeExposed;
    in {
      packages = forAllSystems (system:
        let
          pkgs = import nixpkgs { inherit system; };
        in {
          subtui = pkgs.buildGoModule {
            pname = "subtui";
            version = "unstable-${self.shortRev or self.dirtyShortRev or "dev"}";

            src = lib.cleanSource self;
            vendorHash = "sha256-XtTO9muHlPJu+BHk2+bt7M4tCNGud52cjAswCFpjv2w=";

            proxyVendor = true; 

            nativeBuildInputs = [ pkgs.pkg-config ];
            buildInputs = [ pkgs.mpv ];

            ldflags = [ "-s" "-w" ];
            
            postInstall = ''
              mv $out/bin/SubTUI $out/bin/subtui
            '';

            meta = with lib; {
              description = "A lightweight Subsonic TUI music player with scrobbling and mpv backend";
              homepage = "https://github.com/MattiaPun/SubTUI";
              license = licenses.mit;
              mainProgram = "subtui";
            };
          };

          default = self.packages.${system}.subtui;
        });

      devShells = forAllSystems (system:
        let
          pkgs = import nixpkgs { inherit system; };
        in {
          default = pkgs.mkShell {
            inputsFrom = [ self.packages.${system}.subtui ];
            
            nativeBuildInputs = with pkgs; [
              go
              gopls
              go-tools
            ];
          };
        });
    };
}
