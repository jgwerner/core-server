#!/bin/bash

#!/bin/bash
set -e

VERSION="1.8.1"

DFILE="go$VERSION.linux-amd64.tar.gz"

if [ -d "$HOME/.go" ] || [ -d "$HOME/go" ]; then
    echo "Installation directories already exist. Exiting."
    exit 1
fi

echo "Downloading $DFILE ..."
wget https://storage.googleapis.com/golang/$DFILE -O /tmp/go.tar.gz
if [ $? -ne 0 ]; then
    echo "Download failed! Exiting."
    exit 1
fi

echo "Extracting ..."
tar -C "$HOME" -xzf /tmp/go.tar.gz

echo "Export environment variables ..."
GOROOT="$HOME/go"
PATH="$PATH:$GOROOT/bin"
GOPATH="$HOME/go"
PATH="$PATH:${GOPATH//://bin:}/bin"

mkdir -p "$GOROOT"/{src,pkg,bin}
echo -e "\nGo $VERSION was installed."
rm -f /tmp/go.tar.gz
