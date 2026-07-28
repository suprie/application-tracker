.PHONY: build-rust build-go build clean run web-install web-build web-dev serve

RUST_BINARY = rust/target/release/ats-reader
BIN_DIR = bin

build-rust:
	cargo build --release --manifest-path rust/Cargo.toml
	mkdir -p $(BIN_DIR)
	cp $(RUST_BINARY) $(BIN_DIR)/ats-reader

build-go:
	go build -o $(BIN_DIR)/ats ./cmd/ats

build: build-rust build-go

clean:
	rm -rf $(BIN_DIR)

run: build
	./$(BIN_DIR)/ats $(ARGS)

web-install:
	cd web && npm install

# Builds the Svelte frontend into internal/web/dist (embedded via //go:embed).
web-build:
	cd web && npm run build

web-dev:
	cd web && npm run dev

# Full stack: build the Go binary + frontend, then serve.
serve: build-go web-build
	./$(BIN_DIR)/ats serve
