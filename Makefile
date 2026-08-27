.PHONY: build

go_files = $(wildcard hubtojo/*.go)
build: $(go_files)
	@mkdir -p build
	cd hubtojo && go build -o ../build/hubtojo
