BINARY_NAME=orchestra
RELEASE_DIR=dist
VERSION=v1.4.0
SHELL := /bin/bash

#OS/ARCH/ARM_VERSION
PLATFORMS := \
	darwin/amd64 darwin/arm64 \
	freebsd/386 freebsd/amd64 freebsd/arm64 freebsd/arm/v5 freebsd/arm/v6 freebsd/arm/v7 \
	linux/386 linux/amd64 linux/arm/v5 linux/arm/v6 linux/arm/v7 \
	openbsd/386 openbsd/amd64 openbsd/arm64 openbsd/arm/v5 openbsd/arm/v6 openbsd/arm/v7 openbsd/riscv64 \
	solaris/amd64 \
	windows/386 windows/amd64 windows/arm64

release:
	rm -rf $(RELEASE_DIR) && mkdir -p $(RELEASE_DIR)
	@for platform in $(PLATFORMS); do \
		OS=$$(echo $$platform | cut -d'/' -f1); \
		ARCH=$$(echo $$platform | cut -d'/' -f2); \
		V=$$(echo $$platform | cut -d'/' -f3); \
		\
		GOOS=$$OS GOARCH=$$ARCH GOARM=$${V#v} \
		OUT_NAME=orchestra-$$OS-$$ARCH$$V; \
		\
		EXT=""; [ "$$OS" = "windows" ] && EXT=".exe"; \
		mkdir -p $(RELEASE_DIR)/$$OUT_NAME; \
		go build -ldflags="-s -w" -o $(RELEASE_DIR)/$$OUT_NAME/$(BINARY_NAME)$$EXT .; \
		cp LICENSE $(RELEASE_DIR)/$$OUT_NAME/; \
		\
		tar -C $(RELEASE_DIR) -czf $(RELEASE_DIR)/$$OUT_NAME.tar.gz $$OUT_NAME; \
		rm -rf $(RELEASE_DIR)/$$OUT_NAME; \
		echo "packaged $$OUT_NAME.tar.gz"; \
	done

