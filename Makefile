BINARY=decoreba

build:
	go build -ldflags "-s -w" -o $(BINARY) ./cmd/decoreba

install:
	go install ./cmd/decoreba

run: build
	./$(BINARY)

clean:
	rm -f $(BINARY)

.PHONY: build install run clean
