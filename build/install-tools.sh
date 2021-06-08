#!/bin/sh
# installs tools required for makefile to work

echo "-- Installing dep"
curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh -
echo "-- Installed dep\n"

echo "-- Installing revive"
go get -u github.com/mgechev/revive
echo "-- Installed revive\n"

echo "-- Installing swaggo"
go get -u github.com/swaggo/swag/cmd/swag
echo "-- Installed swaggo\n"
