# Debian package

## Procedure using SPT

```sh
spt create shinokada/tera
# Use amd64 for Architecture
spt open
```

### Structure
Remove all unnecessary dir/files:

```sh
cd /Users/shinichiokada/.cache/spt/pkg/tera_1.0.0-rc.2-1_amd64/usr/share/tera && rm -rf cmd images internal pkg .git .github && rm -f .gitignore .coderabbit.yml CNAME components.test Makefile note-to-ai.md RELEASING.md robots.txt .goreleaser.yaml go.mod go.sum && rmdir /Users/shinichiokada/.cache/spt/pkg/tera_1.0.0-rc.2-1_amd64/usr/share/tera
```

```text
tera_1.0.0-rc.2-1_amd64/
├── DEBIAN/
│   ├── control
│   └── preinst
└── usr/
    ├── bin/
    │   └── tera
    └── share/
        └── doc/
            └── tera/
                └── README.md   (and LICENSE if you added it)
```

DEBIAN/control file
```sh
Package: tera
Version: 1.0.0-rc.2
Architecture: amd64
Maintainer: Shinichi Okada <147320+shinokada@users.noreply.github.com>
Depends: mpv
Section: utils
Priority: optional
Homepage: https://github.com/shinokada/tera
Description: A terminal-based internet radio player powered by Radio Browser
```

DEBIAN/preinst:
```sh
#!/bin/bash
set -e

case "$1" in
    install|upgrade)
        echo "Checking for old versions of tera ..."

        if [ -f "/usr/bin/tera" ]; then
            rm -f "/usr/bin/tera"
            echo "Removed old tera from /usr/bin"
        fi

        if [ -d "/usr/share/tera" ]; then
            rm -rf "/usr/share/tera"
            echo "Removed old tera from /usr/share"
        fi
        ;;
esac

exit 0
```

## chmod
Ensure chmod 755 on the preinst script
```sh
chmod +x path/to/DEBIAN/preinst
```