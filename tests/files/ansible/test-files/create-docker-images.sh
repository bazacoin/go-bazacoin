#!/bin/bash -x

# creates the necessary docker images to run testrunner.sh locally

docker build --tag="bazacoin/cppjit-testrunner" docker-cppjit
docker build --tag="bazacoin/python-testrunner" docker-python
docker build --tag="bazacoin/go-testrunner" docker-go
