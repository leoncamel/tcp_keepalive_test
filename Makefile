
all:build

clean:
	rm -rf dist

build:
	rm -f tcp_test && go build

release:
	goreleaser --snapshot --skip-publish --rm-dist

