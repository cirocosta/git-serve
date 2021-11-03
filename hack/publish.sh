#!/usr/bin/env bash

set -o errexit

checksums_file=$(mktemp)
pushd dist
find . -name "git-serve*" -type f | xargs sha256sum > $checksums_file
popd

gh release create draft-$(date +%s) \
	--draft \
	--notes-file <(echo "<p>sha256sum</p><pre>$(cat $checksums_file)</pre>") \
	./dist/*
