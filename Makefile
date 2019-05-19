build:
	@go build -o fd8-judge

clean:
	@rm -rf fd8-judge

update-deps:
	@rm -rf vendor
	@govendor init
	@govendor add +outside
	@git add vendor

.PHONY: build clean
