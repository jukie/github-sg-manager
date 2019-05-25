BIN             = hookCIDRs
OUTPUT_DIR      = build

export AWS_REGION = us-east-1

.PHONY: help
.DEFAULT_GOAL := help

build: clean ## Build an executable linux binary
	GOOS=linux CGO_ENABLED=0 go build -a -installsuffix cgo -o $(OUTPUT_DIR)/$(BIN) lambda.go
	cd $(OUTPUT_DIR) && zip $(BIN).zip $(BIN)

clean: ## Remove all build artifacts
	$(RM) $(OUTPUT_DIR)/$(BIN).zip
	$(RM) $(OUTPUT_DIR)/$(BIN)

test : ## Integration Testing - ToDo
	go run local.go 
help: ## Display this help message
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_\/-]+:.*?## / {printf "\033[34m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST) | \
		sort | \
		grep -v '#'	