VERSION := $(shell git describe --always --dirty)

help:
	@echo 'make build   - build the passage executable'
	@echo 'make tag     - tag the current HEAD with VERSION'
	@echo 'make archive - create an archive of the current HEAD for VERSION'
	@echo 'make upload  - upload the built archive of VERSION to Amazon S3'
	@echo 'make all     - build, tag, archive and upload VERSION'

version:
	@if [ "$$VERSION" = "" ]; then echo "VERSION not set"; exit 1; fi

passage_linux:
	@echo 'Unsupported architecture: $@'

passage_darwin:
	GOOS=darwin GOARCH=amd64 go build -ldflags="-X main.version=${VERSION}" -o passage_darwin passage.go

passage_all: passage_linux passage_darwin

passage: passage.go
	go build -ldflags="-X main.version=${VERSION}" $^

build: passage

tag: version
	git tag -s $$VERSION -m "$$VERSION release"

archive: passage-${VERSION}.zip

passage-${VERSION}.zip: passage
	git archive -o $@ HEAD
	zip $@ passage

upload: passage-${VERSION}.zip
	aws s3 cp "passage-${VERSION}.zip" s3://wincent/passage/releases/passage-${VERSION}.zip --acl public-read

all: tag build archive upload

.PHONY: clean
clean:
	rm -f passage passage-*.zip
	rm -f passage_*
