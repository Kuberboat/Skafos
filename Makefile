BUILD_DIR = ./out/bin
CMD_SOURCE_DIRS = cmd
SOURCE_DIRS = cmd pkg
SKPROXY_SRC = ./cmd/skproxy/skproxy.go
SKPROXY_OBJ = skproxy
SCRIPTS_DIR = ./scripts
SHFMT_FLAG = shfmt
XARGS_FLAG = xargs

$(shell mkdir -p $(BUILD_DIR))

export GO111MODULE := on
export GOPROXY := https://mirrors.aliyun.com/goproxy/,direct

all: skproxy

skproxy: $(SKPROXY_SRC)
	@go build -o $(BUILD_DIR)/$(SKPROXY_OBJ) $(SKPROXY_SRC)

.PHONY: image
image:
	$(SCRIPTS_DIR)/skproxy/build-image.sh

.PHONY: fmt
fmt:
	@gofmt -s -w $(SOURCE_DIRS)
	$(SHFMT_FLAG) -f . | $(XARGS_FLAG) $(SHFMT_FLAG) -w

.PHONY: clean
clean:
	rm -rf $(BUILD_DIR)
