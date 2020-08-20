example-image:
	docker build -t io-example ./example

volume:
	docker volume create io-example

example: example-image volume
	docker run -v io-example:/mnt/test -it io-example /main-app

example-inject:debug-toda
	cat ./io-inject-example.json|sudo ./target/debug/toda --path /mnt/test --pid $$(pgrep main-app) --verbose trace

debug-toda:
	RUSTFLAGS="-Z relro-level=full" cargo build