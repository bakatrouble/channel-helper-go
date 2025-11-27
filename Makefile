GO_FILES = $(shell find . -name '*.go')
RUST_FILES = $(shell find lib/imagehash -name '*')
MINIAPP_FILES = $(shell find miniapp/ -not -path 'miniapp/dist/*' -not -path 'miniapp/node_modules/*' -name '*')

channel-helper-go: $(GO_FILES) ./lib/libimagehash.so ./miniapp/dist/index.html
	go build -ldflags="-r lib"

./lib/libimagehash.so: $(RUST_FILES)
	cd lib/imagehash && cargo build --release
	cp lib/imagehash/target/release/libimagehash.so lib

./miniapp/dist/index.html: $(MINIAPP_FILES)
	cd miniapp && pnpm i && pnpm build
