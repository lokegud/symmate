{ pkgs ? import <nixpkgs> {} }:
pkgs.mkShell {
  packages = with pkgs; [
    bun nodejs_22
    rustc cargo pkg-config
    webkitgtk_4_1 librsvg gtk3 cairo gdk-pixbuf glib
    openssl libsoup_3
  ];
  shellHook = ''
    export PKG_CONFIG_PATH="${pkgs.webkitgtk_4_1.dev}/lib/pkgconfig:${pkgs.libsoup_3.dev}/lib/pkgconfig:$PKG_CONFIG_PATH"
  '';
}
