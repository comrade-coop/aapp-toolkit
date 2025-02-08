#!/bin/bash

apt-get update -yq
apt-get install -yq build-essential
apt-get install -yq libcurl4-openssl-dev
apt-get install -yq libjsoncpp-dev
apt-get install -yq libboost-all-dev
apt-get install -yq cmake
apt-get install -yq nlohmann-json3-dev
wget https://packages.microsoft.com/repos/azurecore/pool/main/a/azguestattestation1/azguestattestation1_1.0.5_amd64.deb
dpkg -i azguestattestation1_1.0.5_amd64.deb