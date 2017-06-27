+ACM- adjusted from https://github.com/bazacoin/cpp-bazacoin/blob/develop/docker/Dockerfile
FROM ubuntu:14.04

ENV DEBIAN+AF8-FRONTEND noninteractive
RUN apt-get update
RUN apt-get upgrade -y

+ACM- Bazacoin dependencies
RUN apt-get install -qy build-essential g4.8 git cmake libboost-all-dev libcurl4-openssl-dev wget
RUN apt-get install -qy automake unzip libgmp-dev libtool libleveldb-dev yasm libminiupnpc-dev libreadline-dev scons
RUN apt-get install -qy libjsoncpp-dev libargtable2-dev

+ACM- NCurses based GUI (not optional though for a successful compilation, see https://github.com/bazacoin/cpp-bazacoin/issues/452 )
RUN apt-get install -qy libncurses5-dev

+ACM- Qt-based GUI
+ACM- RUN apt-get install -qy qtbase5-dev qt5-default qtdeclarative5-dev libqt5webkit5-dev

+ACM- Bazacoin PPA
RUN apt-get install -qy software-properties-common
RUN add-apt-repository ppa:bazacoin/bazacoin
RUN apt-get update
RUN apt-get install -qy libcryptopp-dev libjson-rpc-cpp-dev

+ACM- Build Bazacoin (HEADLESS)
RUN git clone --depth+AD0-1 --branch develop https://github.com/bazacoin/cpp-bazacoin
RUN mkdir -p cpp-bazacoin/build
RUN cd cpp-bazacoin/build +ACYAJg- cmake .. -DCMAKE+AF8-BUILD+AF8-TYPE+AD0-Release -DHEADLESS+AD0-1 +ACYAJg- make -j +ACQ-(cat /proc/cpuinfo +AHw- grep processor +AHw- wc -l) +ACYAJg- make install
RUN ldconfig

ENTRYPOINT +AFsAIg-/cpp-bazacoin/build/test/createRandomTest+ACIAXQ-

