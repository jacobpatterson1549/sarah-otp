.PHONY: all serve doc clean

BUILD_DIR := build
PKG_WASM=github.com/jacobpatterson1549/sarah-otp/go/cmd/ui
PKG_SERVER=github.com/jacobpatterson1549/sarah-otp/go/cmd/server
OBJ_TEST_WASM=test_wasm.txt
OBJ_TEST_SERVER=test_server.txt
OBJ_WASM=main.wasm
OBJ_SERVER=main
OBJ_RESOURCES=resources
OBJ_VERSION=version
OBJ_WASM_EXEC_JS=wasm_exec.js
OBJ_GO_JS_WASM_EXEC=go_js_wasm_exec
SERVER_DIR := go/cmd/server
GO_LIST := go list ./...
GO_TEST := go test ./... -cover# -race
GO_BUILD := go build# -race
GO_DOC := go doc
GO_WASM_ARGS := GOOS=js GOARCH=wasm
GO_ARGS :=
GO_WASM_PATH := $(shell go env GOROOT)/lib/wasm
SRC_GO := $(shell find go -name *.go)
SRC_RESOURCES := $(wildcard resources/*) \
	$(wildcard resources/http/*)

all: $(BUILD_DIR)/$(OBJ_SERVER)

$(BUILD_DIR):
	mkdir $@

$(SERVER_DIR)/$(BUILD_DIR):
	mkdir $@

$(SERVER_DIR)/$(BUILD_DIR)/$(OBJ_WASM_EXEC_JS): | $(SERVER_DIR)/$(BUILD_DIR)
	cp \
		$(GO_WASM_PATH)/$(@F) \
		$@

$(SERVER_DIR)/$(OBJ_RESOURCES): $(SRC_RESOURCES)
	cp -R \
		$(@F) \
		$@

$(SERVER_DIR)/$(BUILD_DIR)/$(OBJ_WASM): $(SRC_GO) $(SERVER_DIR)/$(BUILD_DIR)/$(OBJ_WASM_EXEC_JS) | $(SERVER_DIR)/$(BUILD_DIR)
	$(GO_WASM_ARGS) $(GO_BUILD) \
		-o $@ \
		$(PKG_WASM)

$(SERVER_DIR)/$(BUILD_DIR)/$(OBJ_VERSION): $(SERVER_DIR)/$(BUILD_DIR)/$(OBJ_WASM) $(SERVER_DIR)/$(OBJ_RESOURCES) | $(SERVER_DIR)/$(BUILD_DIR)
	find go \
			-not -path $@ \
			-not -path $< \
			-type f \
		| xargs cat \
		| md5sum \
		| cut -c -32 \
		| tee $@ \
		| xargs echo $(@F)

$(BUILD_DIR)/$(OBJ_TEST_WASM): $(SERVER_DIR)/$(BUILD_DIR)/$(OBJ_WASM_EXEC_JS) $(SERVER_DIR)/$(BUILD_DIR)/$(OBJ_WASM_EXEC_JS)  $(SERVER_DIR)/$(BUILD_DIR)/$(OBJ_VERSION) | $(BUILD_DIR)
	$(GO_WASM_ARGS) $(GO_TEST) \
		-exec=$(GO_WASM_PATH)/$(OBJ_GO_JS_WASM_EXEC) \
		| tee $@

$(BUILD_DIR)/$(OBJ_TEST_SERVER): $(SERVER_DIR)/$(BUILD_DIR)/$(OBJ_WASM_EXEC_JS) $(SERVER_DIR)/$(BUILD_DIR)/$(OBJ_WASM_EXEC_JS)  $(SERVER_DIR)/$(BUILD_DIR)/$(OBJ_VERSION) | $(BUILD_DIR)
	$(GO_ARGS) $(GO_TEST) \
		| tee $@

$(BUILD_DIR)/$(OBJ_SERVER): $(BUILD_DIR)/$(OBJ_TEST_WASM) $(BUILD_DIR)/$(OBJ_TEST_SERVER) | $(BUILD_DIR)
	$(GO_ARGS) $(GO_BUILD) \
		-o $@ \
		$(PKG_SERVER)

serve: $(BUILD_DIR)/$(OBJ_SERVER)
	export $(shell grep -s -v '^#' .env | xargs) \
		&& ./$<

doc:
	$(GO_DOC) -u -http

clean:
	rm -rf $(BUILD_DIR) $(SERVER_DIR)/$(OBJ_RESOURCES) $(SERVER_DIR)/$(BUILD_DIR)
