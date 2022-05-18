BUILD_DIR = ./out/bin
CMD_SOURCE_DIRS = cmd
SOURCE_DIRS = cmd pkg
SKPROXY_SRC = ./cmd/skproxy/skproxy.go
SKPROXY_OBJ = skproxy
SKCTL_SRC = ./cmd/skctl/skctl.go
SKCTL_OBJ = skctl
SKAGENT_SRC = ./cmd/skagent/skagent.go
SKAGENT_OBJ = skagent
SKPILOT_SRC = ./cmd/skpilot/skpilot.go
SKPILOT_OBJ = skpilot
PROTO_GEN_DIR = ./pkg/proto
PROTO_SCRIPT = kuberboat/scripts/proto_gen.sh
SCRIPTS_DIR = ./scripts
SHFMT_FLAG = shfmt
XARGS_FLAG = xargs

$(shell mkdir -p $(BUILD_DIR))

export GO111MODULE := on
export GOPROXY := https://mirrors.aliyun.com/goproxy/,direct

all: proto skproxy skctl skpilot skagent

skproxy: $(SKPROXY_SRC)
	@go build -o $(BUILD_DIR)/$(SKPROXY_OBJ) $(SKPROXY_SRC)

.PHONY: image
image:
	$(SCRIPTS_DIR)/skproxy/build-image.sh

skctl: $(SKCTL_SRC)
	@go build -o $(BUILD_DIR)/$(SKCTL_OBJ) $(SKCTL_SRC)

skpilot: $(SKPILOT_SRC)
	@go build -o $(BUILD_DIR)/$(SKPILOT_OBJ) $(SKPILOT_SRC)

skagent: $(SKAGENT_SRC)
	@go build -o $(BUILD_DIR)/$(SKAGENT_OBJ) $(SKAGENT_SRC)

.PHONY: proto
proto:
	rm -rf $(PROTO_GEN_DIR)
	./$(PROTO_SCRIPT)

.PHONY: fmt
fmt:
	@gofmt -s -w $(SOURCE_DIRS)
	$(SHFMT_FLAG) -f . | $(XARGS_FLAG) $(SHFMT_FLAG) -w

.PHONY: clean
clean:
	rm -rf $(BUILD_DIR)
