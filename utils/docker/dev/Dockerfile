# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

FROM hyperledger/fabric-private-chaincode-base:latest

ARG GO_VERSION=go1.14
ARG GO_TAR=$GO_VERSION.linux-amd64.tar.gz
ARG NANOPB_VERSION=0.3.9.2
ARG FABRIC_REPO=https://github.com/hyperledger/fabric.git
ARG FABRIC_VERSION=v2.0.1
ARG FPC_REPO_URL=https://github.com/hyperledger-labs/fabric-private-chaincode.git
ARG FPC_REPO_BRANCH=master
ARG OPENSSL_VERSION=1.1.0j
ARG SGXSSL_VERSION=v2.4.1

ARG APT_ADD_PKGS=

ARG GIT_USER_NAME=tester
ARG GIT_USER_EMAIL=tester@fpc

ENV NANOPB_VERSION=${NANOPB_VERSION}
ENV FABRIC_VERSION=${FABRIC_VERSION}
ENV OPENSSL_VERSION=${OPENSSL_VERSION}
ENV SGXSSL_VERSION=${SGXSSL_VERSION}


WORKDIR /tmp

RUN git config --global user.name $GIT_USER_NAME \
 && git config --global user.email $GIT_USER_EMAIL

# we need docker commands (even though we use docker daemon from "outside")
RUN apt-get update \
 && apt-get install -y \
    docker.io \
    clang-format \
    jq \
    $APT_ADD_PKGS

# Install go
RUN wget https://dl.google.com/go/$GO_TAR \
 && tar -xf $GO_TAR \
 && mv go /usr/local \
 && rm $GO_TAR \
 && mkdir -p /project

ENV GOROOT=/usr/local/go
ENV GOPATH=/project
ENV PATH=$GOPATH/bin:$GOROOT/bin:$PATH

# Install SGX SSL
ENV SGX_SSL /opt/intel/sgxssl
RUN \
    OPENSSL_MAJOR_VERSION=$(echo ${OPENSSL_VERSION} | sed 's/\([^0-9.]\)*//g') \
 && wget https://www.openssl.org/source/old/${OPENSSL_MAJOR_VERSION}/openssl-${OPENSSL_VERSION}.tar.gz \
 && git clone  --branch ${SGXSSL_VERSION} https://github.com/intel/intel-sgx-ssl.git \
 && . $SGX_SDK/environment \
 && (cd intel-sgx-ssl/openssl_source; mv /tmp/openssl-${OPENSSL_VERSION}.tar.gz . ) \
 && (cd intel-sgx-ssl/Linux; make SGX_MODE=SIM DESTDIR=$SGX_SSL all test ) \
 # Note: in docker build we cannot run with SGX_MODE=HW, so run as SIM.
 # However, compiled libraries can still be used in both modes!
 && (cd intel-sgx-ssl/Linux; make install ) \
 && rm -rf /tmp/intel-sgx-ssl

# Install nanopb
ENV NANOPB_PATH=/usr/local/nanopb/
RUN git clone https://github.com/nanopb/nanopb.git $NANOPB_PATH \
 && cd $NANOPB_PATH \
 && git checkout nanopb-$NANOPB_VERSION \
 && cd generator/proto \
 && make

# Go tools we need
RUN go get golang.org/x/tools/cmd/goimports

# Get Fabric
ENV FABRIC_PATH=$GOPATH/src/github.com/hyperledger/fabric
RUN git clone $FABRIC_REPO $FABRIC_PATH \
# Note: could add --single-branch to below to speed-up and keep size smaller. But for now for a dev-image better keep complete repo
 && cd $FABRIC_PATH \
 && git checkout tags/$FABRIC_VERSION

# Get FPC
ENV FPC_PATH=$GOPATH/src/github.com/hyperledger-labs/fabric-private-chaincode
RUN git clone --recursive --branch $FPC_REPO_BRANCH $FPC_REPO_URL $FPC_PATH
# Note: could add --single-branch to below to speed-up and keep size smaller. But for now for a dev-image better keep complete repo

WORKDIR $FPC_PATH
