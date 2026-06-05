.PHONY: build-rust build-go build clean run

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
