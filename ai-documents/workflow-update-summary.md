# GitHub Workflow Update Summary

## Changes Made

Updated `.github/workflows/test.yml` from Bash/BATS testing to Go testing.

## Old Workflow (Bash)
- Used BATS (Bash Automated Testing System)
- Installed BATS via apt-get (Ubuntu) or brew (macOS)
- Ran `bats . --formatter tap` in tests directory

## New Workflow (Go)

### Three Jobs

#### 1. Test Job
- **Matrix Strategy**: Tests across Ubuntu/macOS with Go 1.21 and 1.22
- **Steps**:
  - Checkout code
  - Setup Go using `actions/setup-go@v5`
  - Download and verify Go modules
  - Run tests with race detection: `go test -v -race -coverprofile=coverage.out ./...`
  - Generate HTML coverage report
  - Upload coverage artifacts
  - Display coverage percentage

#### 2. Build Job
- **Matrix Strategy**: Build on Ubuntu and macOS
- **Steps**:
  - Build binary: `go build -v -o tera ./cmd/tera`
  - Verify binary can execute
  - Show binary size

#### 3. Lint Job
- **Runs on**: Ubuntu only (linting once is sufficient)
- **Uses**: `golangci-lint-action@v4`
- **Timeout**: 5 minutes

## Key Features

✅ **Cross-platform testing** - Ubuntu and macOS  
✅ **Multi-version Go support** - Tests on Go 1.21 and 1.22  
✅ **Race detection** - Catches concurrency bugs  
✅ **Coverage reporting** - Generates and uploads HTML reports  
✅ **Build verification** - Ensures binary compiles on all platforms  
✅ **Linting** - Code quality checks with golangci-lint  
✅ **Fail-fast disabled** - All matrix jobs run even if one fails  

## Go Flags Explained

- `-v`: Verbose output (show all tests)
- `-race`: Enable race detector for concurrency bugs
- `-coverprofile=coverage.out`: Generate coverage profile
- `./...`: Run tests in all subdirectories

## Optional Enhancement

The workflow includes a commented section to fail if coverage drops below 50%. Uncomment to enforce minimum coverage:

```yaml
if (( $(echo "$coverage < 50" | bc -l) )); then
  echo "Coverage is below 50%"
  exit 1
fi
```

## Artifacts

Each test run uploads coverage reports as artifacts:
- `coverage-ubuntu-latest-go1.21`
- `coverage-ubuntu-latest-go1.22`
- `coverage-macos-latest-go1.21`
- `coverage-macos-latest-go1.22`
