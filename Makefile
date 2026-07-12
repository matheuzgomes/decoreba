BINARY=decoreba
DESKTOP_BIN=cmd/decoreba-desktop/build/bin/decoreba-desktop

build:
	go build -ldflags "-s -w" -o $(BINARY) ./cmd/decoreba

build-wails:
	cd cmd/decoreba-desktop && wails build

install:
	go install ./cmd/decoreba

run: build
	./$(BINARY)

run-desktop: build-wails
	./$(DESKTOP_BIN)

clean:
	rm -f $(BINARY)

.PHONY: build build-wails install run run-desktop clean
