#!/bin/sh
rm -f ev
go get github.com/cenkalti/rpc2
go build .
echo "done"