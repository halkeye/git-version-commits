install:
	go install ./...

all:
	mkdir -p bin
	for binary in confluence-poster git-release-info release-info-confluence; do \
		env GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o bin/$${binary}-darwin-amd64 $${binary}/main.go; \
		env GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o bin/$${binary}-linux-amd64 $${binary}/main.go; \
		env GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o bin/$${binary}-windows-amd64 $${binary}/main.go; \
	done

