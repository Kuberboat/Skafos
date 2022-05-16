BUILD_DIR = ./out/bin
SOURCE_DIRS = cmd pkg
SHFMT_FLAG = shfmt
XARGS_FLAG = xargs

$(shell mkdir -p $(BUILD_DIR))

export GO111MODULE := on
export GOPROXY := https://mirrors.aliyun.com/goproxy/,direct

.PHONY: fmt
fmt:
	@gofmt -s -w $(SOURCE_DIRS)
	$(SHFMT_FLAG) -f . | $(XARGS_FLAG) $(SHFMT_FLAG) -w

.PHONY: clean
clean:
	rm -rf $(BUILD_DIR)
