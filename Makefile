BINARY=decoreba
DESKTOP_DIR=cmd/decoreba-desktop
DESKTOP_BIN=$(DESKTOP_DIR)/build/bin/decoreba-desktop

build:
	go build -ldflags "-s -w" -o $(BINARY) ./cmd/decoreba

build-wails:
	cd $(DESKTOP_DIR) && wails build

install:
	go install ./cmd/decoreba

run: build
	./$(BINARY)

run-desktop: build-wails
	./$(DESKTOP_BIN)

clean:
	rm -f $(BINARY)

.PHONY: build build-wails install run run-desktop clean
