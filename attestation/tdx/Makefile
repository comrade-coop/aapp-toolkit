cbindings:
	mkdir -p c
	cargo build --features clib --release
	cp target/release/libtdx.so c/libtdx.so

.PHONY: cbindings