#!/bin/bash

go get github.com/ibmdb/go_ibm_db
export DB2HOME=$GOPATH/src/github.com/ibmdb/go_ibm_db/installer
cur="$PWD"
cd $DB2HOME && go run setup.go
export CGO_CFLAGS=-I$DB2HOME/include
export CGO_LDFLAGS=-L$DB2HOME/lib
export DYLD_LIBRARY_PATH=$DYLD_LIBRARY_PATH:$DB2HOME/clidriver/lib
cd $cur
go test -db=go_ibm_db -tags=db2 -conn_str="HOSTNAME=localhost;DATABASE=testdb;PORT=50000;UID=db2inst1;PWD=password"