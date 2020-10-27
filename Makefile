.PHONY: all test-wasm test serve serve-tcp clean

BUILD_DIR := build
GO_LIST := go list ./...
GO_TEST := go test --cover # -race
GO_BUILD := go build # -race
GO_WASM_ARGS := GOOS=js GOARCH=wasm
GO_ARGS :=
GO_WASM_PATH := $(shell go env GOROOT)/misc/wasm
LINK := ln -fs
OBJS := $(addprefix $(BUILD_DIR)/,main.wasm main version wasm_exec.js resources)

all: $(OBJS)

test-wasm:
	$(GO_WASM_ARGS) $(GO_LIST) | grep ui \
		| $(GO_WASM_ARGS) xargs $(GO_TEST) \
			-exec=$(GO_WASM_PATH)/go_js_wasm_exec

test:
	$(GO_LIST) | grep -v ui \
		| $(GO_ARGS) xargs $(GO_TEST)

$(BUILD_DIR):
	mkdir $(BUILD_DIR)

$(BUILD_DIR)/wasm_exec.js: | $(BUILD_DIR)
	$(LINK) \
		$(GO_WASM_PATH)/$(@F) \
		$@

$(BUILD_DIR)/resources: | $(BUILD_DIR)
	$(LINK) \
		$(PWD)/$(@F) \
		$@

$(BUILD_DIR)/version: | $(BUILD_DIR)
	find . \
			-mindepth 2 \
			-path "*/.*" -prune -o \
			-path "./$(BUILD_DIR)/*" -prune -o \
			-type f \
			-print \
		| xargs tar -c \
		| md5sum \
		| cut -c -32 \
		| tee $@ \
		| xargs echo version

$(BUILD_DIR)/main.wasm: test-wasm | $(BUILD_DIR)
	$(GO_WASM_ARGS) $(GO_LIST) | grep cmd/ui \
		| $(GO_WASM_ARGS) xargs $(GO_BUILD) \
			-o $@

$(BUILD_DIR)/main: test | $(BUILD_DIR)
	$(GO_LIST) | grep cmd/server \
		| $(GO_ARGS) xargs $(GO_BUILD) \
			-o $@

serve: all
	export $(shell grep -s -v '^#' .env | xargs) \
		&& cd $(BUILD_DIR) \
		&& ./main

serve-tcp: all
	sudo setcap 'cap_net_bind_service=+ep' $(BUILD_DIR)/main
	export $(shell grep -s -v '^#' .env | xargs \
			| xargs -I {} echo "{} HTTP_PORT=80 HTTPS_PORT=443") \
		&& cd $(BUILD_DIR) \
		&& sudo -E ./main

clean:
	rm -rf $(BUILD_DIR)
