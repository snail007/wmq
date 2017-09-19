#!/bin/bash
VER="1.5"
RELEASE="wmq_linux_amd64-${VER}"
rm -rf ${RELEASE}
mkdir ${RELEASE}
set CGO_ENABLED=0
GOOS=linux GOARCH=amd64 go build 
mv wmq ${RELEASE}/
cp config.toml ${RELEASE}/
tar zcfv "${RELEASE}.tar.gz" "${RELEASE}"
rm -rf wmq_linux_amd64*