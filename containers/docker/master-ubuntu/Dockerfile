FROM ubuntu:xenial

RUN \
  apt-get update && apt-get upgrade -q -y && \
  apt-get install -y --no-install-recommends golang git make gcc libc-dev ca-certificates && \
  git clone --depth 1 --branch release/1.5 https://github.com/bazacoin/go-bazacoin && \
  (cd go-bazacoin && make geth) && \
  cp go-bazacoin/build/bin/geth /geth && \
  apt-get remove -y golang git make gcc libc-dev && apt autoremove -y && apt-get clean && \
  rm -rf /go-bazacoin

EXPOSE 8545
EXPOSE 30303

ENTRYPOINT ["/geth"]
