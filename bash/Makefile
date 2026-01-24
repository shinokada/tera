# Makefile for TERA tests

.PHONY: test quick-test install-bats help clean

# Default target
help:
	@echo "TERA Test Suite"
	@echo ""
	@echo "Available targets:"
	@echo "  make test         - Run all tests"
	@echo "  make quick-test   - Run critical tests only (fast)"
	@echo "  make install-bats - Install BATS testing framework"
	@echo "  make clean        - Clean up test artifacts"
	@echo "  make help         - Show this help message"

# Run all tests
test:
	@echo "Running all tests..."
	@cd tests && bats .

# Run quick tests
quick-test:
	@echo "Running critical tests..."
	@cd tests && ./quick_test.sh

# Install BATS (macOS with Homebrew)
install-bats:
	@echo "Installing BATS..."
	@if command -v brew >/dev/null 2>&1; then \
		brew install bats-core; \
	elif command -v apt-get >/dev/null 2>&1; then \
		sudo apt-get update && sudo apt-get install -y bats; \
	else \
		echo "Please install BATS manually: https://github.com/bats-core/bats-core"; \
	fi

# Clean up
clean:
	@echo "Cleaning up test artifacts..."
	@rm -f tests/*.tmp
	@rm -rf ~/.cache/tera/test_*
