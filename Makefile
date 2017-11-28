install:
	go install ./...

all:
	mkdir -p bin
	for binary in confluence-poster git-release-info release-info-confluence; do \
		for OS in darwin linux windows; do \
			env CGO_ENABLED=0 GOOS=$$OS GOARCH=amd64 go build -a -installsuffix cgo -o bin/$${binary}-$${OS}-amd64 $${binary}/main.go; \
		done; \
	done

docker:
	docker build -t halkeye/git-version-commits .
