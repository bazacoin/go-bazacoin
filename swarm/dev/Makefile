.PHONY: build cluster test

default: build

build:
	go build -o bin/swarm    github.com/bazacoin/go-bazacoin/cmd/swarm
	go build -o bin/geth     github.com/bazacoin/go-bazacoin/cmd/geth
	go build -o bin/bootnode github.com/bazacoin/go-bazacoin/cmd/bootnode

cluster: build
	scripts/boot-cluster.sh

test:
	go test -v github.com/bazacoin/go-bazacoin/swarm/...
