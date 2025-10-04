GO_FILES = $(shell find . -name '*.go')
RUST_FILES = $(shell find lib/imagehash -name '*')

channel-helper-go: $(GO_FILES) ./lib/libimagehash.so
	go build -ldflags="-r lib"

./lib/libimagehash.so: $(RUST_FILES)
	cd lib/imagehash && cargo build --release
	cp lib/imagehash/target/release/libimagehash.so lib
