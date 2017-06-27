#!/bin/sh

set -e

if [ ! -f "build/env.sh" ]; then
    echo "$0 must be run from the root of the repository."
    exit 2
fi

# Create fake Go workspace if it doesn't exist yet.
workspace="$PWD/build/_workspace"
root="$PWD"
bzcdir="$workspace/src/github.com/bazacoin"
if [ ! -L "$bzcdir/go-bazacoin" ]; then
    mkdir -p "$bzcdir"
    cd "$bzcdir"
    ln -s ../../../../../. go-bazacoin
    cd "$root"
fi

# Set up the environment to use the workspace.
GOPATH="$workspace"
export GOPATH

# Run the command inside the workspace.
cd "$bzcdir/go-bazacoin"
PWD="$bzcdir/go-bazacoin"

# Launch the arguments with the configured environment.
exec "$@"
