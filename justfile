app       := "harvester"
org       := "github.com/sfborg/"
version   := `git describe --tags`
ver       := `git describe --tags --abbrev=0`
date      := `date -u '+%Y-%m-%d_%H:%M:%S'`
no_c      := "CGO_ENABLED=0"
x86       := "GOARCH=amd64"
arm       := "GOARCH=arm64"
linux     := "GOOS=linux"
mac       := "GOOS=darwin"
win       := "GOOS=windows"
flags_ld  := "-ldflags '-X " + org + app + "/pkg.Build=" + date + \
             " -X " + org + app + "/pkg.Version=" + version + "'"
flags_rel := "-trimpath -ldflags '-s -w -X " + org + app + \
             "/pkg.Build=" + date + "'"

# Default recipe - install to ~/go/bin
default: install

# Run all tests
test:
    go test -count=1 ./...

# Build the binary (development build with timestamp and git version)
build: peg
    @mkdir -p bin
    {{no_c}} go build {{flags_ld}} -o bin/{{app}}
    @echo "✅ {{app}} built to bin/{{app}}"

# Build release binary (uses version.go for Version, timestamp for Build)
build-release: peg
    @mkdir -p bin
    {{no_c}} go build {{flags_rel}} -o bin/{{app}}
    @echo "✅ {{app}} release binary built to bin/{{app}}"

# Install to ~/go/bin (development build with timestamp and git version)
install: peg
    {{no_c}} go install {{flags_ld}}
    @echo "✅ {{app}} installed to ~/go/bin/{{app}}"

# Build releases for multiple platforms
release: peg
    @echo "Building releases for Linux, Mac (Intel), Mac (ARM), Windows"

    @mkdir -p bin/releases

    {{no_c}} {{linux}} {{x86}} go build {{flags_rel}} -o bin/releases/{{app}}
    tar zcvf bin/releases/{{app}}-{{ver}}-linux.tar.gz -C bin/releases {{app}}
    rm bin/releases/{{app}}

    {{no_c}} {{mac}} {{x86}} go build {{flags_rel}} -o bin/releases/{{app}}
    tar zcvf bin/releases/{{app}}-{{ver}}-mac-amd64.tar.gz -C bin/releases {{app}}
    rm bin/releases/{{app}}

    {{no_c}} {{mac}} {{arm}} go build {{flags_rel}} -o bin/releases/{{app}}
    tar zcvf bin/releases/{{app}}-{{ver}}-mac-arm64.tar.gz -C bin/releases {{app}}
    rm bin/releases/{{app}}

    {{no_c}} {{win}} {{x86}} go build {{flags_rel}} -o bin/releases/{{app}}.exe
    cd bin/releases && zip -9 {{app}}-{{ver}}-win-64.zip APP.exe
    rm bin/releases/APP.exe

    @echo "✅ Release binaries created in bin/releases/"

# Clean build artifacts
clean:
    rm -rf bin coverage.out coverage.html
    go clean
    @echo "✅ Cleaned build artifacts"

# Format all Go code
fmt:
    go fmt ./...

# Run linter (requires golangci-lint)
lint:
    golangci-lint run

# Tidy dependencies
tidy:
    go mod tidy

# Generate PEG parser from grammar
peg:
    cd internal/sources/wikisp/wsparser && go run github.com/pointlander/peg name.peg
    @echo "✅ PEG parser generated for wikisp"

# Verify project builds and all tests pass
verify: fmt tidy test build
    @echo "✅ Verification complete: code formatted, dependencies tidied, tests passing, build successful"
