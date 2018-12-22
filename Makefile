programs = confluence-poster git-release-info release-info-confluence

app: deps
	mkdir -p bin
	for binary in $(programs); do \
		go build -o bin/$${binary} $${binary}/main.go; \
	done

install:
	go install ./...

deps:
	go get ./...

all: deps
	mkdir -p bin
	for binary in $(programs); do \
		for OS in darwin linux windows; do \
			env CGO_ENABLED=0 GOOS=$$OS GOARCH=amd64 go build -a -installsuffix cgo -o bin/$${binary}-$${OS}-amd64 $${binary}/main.go; \
		done; \
	done

docker:
	docker build -t halkeye/git-version-commits .

godocker:
	 godockerize build -t git-version-commits:latest github.com/halkeye/git-version-commits/{git-release-info,confluence-poster,release-info-confluence}
