export GRROOT="$PWD"
export GOPATH=${GOPATH}:$GRROOT
export PATH=$PATH:$GRROOT/bin
export GOPATH=${GOPATH#:}
