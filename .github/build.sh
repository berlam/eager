#!/bin/sh

cd ${GITHUB_WORKSPACE:-.}

mkdir -p .dist

go run main.go --version | cut -c 9- > .version

for GOOS in darwin linux windows; do
	SUFFIX=`[ $GOOS = "windows" ] && echo ".exe"`
	for GOARCH in 386 amd64; do
		TARGET=.release/$GOOS/$GOARCH/eager$SUFFIX
		go build -ldflags="-s -w" -v -o $TARGET
		tar --transform 's/.*\///g' -czvf .dist/eager-$GOOS-$GOARCH.tar.gz $TARGET README.md LICENSE
	done
done
