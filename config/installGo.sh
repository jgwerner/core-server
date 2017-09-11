#!/bin/bash
set -e

mkdir $HOME/bin
curl -sL -o $HOME/bin/gimme https://raw.githubusercontent.com/travis-ci/gimme/master/gimme
chmod +x $HOME/bin/gimme
echo "Updating gimme"
GIMME_OUTPUT="$(gimme 1.9 | tee -a $HOME/.bashrc)" && eval "$GIMME_OUTPUT"
export GOPATH=$HOME/gopath
export PATH=$HOME/gopath/bin:$PATH
mkdir -p $HOME/gopath/src/github.com/3Blades/core-server
rsync -az ${TRAVIS_BUILD_DIR}/ $HOME/gopath/src/github.com/3Blades/core-server/
export TRAVIS_BUILD_DIR=$HOME/gopath/src/github.com/3Blades/core-server
cd $HOME/gopath/src/github.com/3Blades/core-server
gimme version
go version
go env
