# List available recipes
default:
    @just --list

# Build the binary
[group: 'build']
build:
    go build -v ./...

# Run all tests
[group: 'test']
test:
    go test -v ./...

# Run tests with coverage
[group: 'test']
test-coverage:
    go test -coverprofile=coverage.out ./...
    go tool cover -func=coverage.out

# Run tests with coverage and open HTML report
[group: 'test']
test-coverage-html: test-coverage
    go tool cover -html=coverage.out

# Format code
[group: 'quality']
fmt:
    go fmt ./...

# Run static analysis
[group: 'quality']
vet:
    go vet ./...

# Tidy module dependencies
[group: 'quality']
tidy:
    go mod tidy

# Run all checks (fmt, vet, test)
[group: 'quality']
check: fmt vet test

# Generate the binary locally
[group: 'build']
bin:
    go build -o ghat .

# Build a snapshot release locally with goreleaser
[group: 'build']
snapshot:
    goreleaser build --snapshot --clean

# Run the binary with arguments
[group: 'dev']
run *ARGS:
    go run . {{ARGS}}

# Clean build artifacts
[group: 'dev']
clean:
    rm -f ghat coverage.out
    rm -rf dist/
