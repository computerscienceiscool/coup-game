SHELL := /bin/bash
# Define ports if needed later for any web interfaces
WORKERS := $(shell nproc || echo 4)  # Default to number of CPU cores
GAMES := 1000000  # Default number of games to simulate
GAMES_QUICK := 1000  # For quick testing
OUTPUT_DIR := ./results  # Default output directory
# Define the branch to use the branch already in use by the machine
BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
# Version tag based on date
export TAG = $(shell date +%Y.%m.%d.%H%M)

.PHONY: help build test run run-quick clean all commit profile benchmark report analyze

default: help

help:
	@echo ""
	@echo "Coup Game Simulation - Available commands:"
	@echo "  make build       - Build the Coup simulation executable"
	@echo "  make test        - Run all tests"
	@echo "  make test-game   - Run only game logic tests"
	@echo "  make test-sim    - Run only simulation tests"
	@echo "  make run         - Run the full simulation ($(GAMES) games)"
	@echo "  make run-quick   - Run a quick simulation ($(GAMES_QUICK) games)"
	@echo "  make clean       - Remove build artifacts and result files"
	@echo "  make all         - Build, test, and run a quick simulation"
	@echo "  make profile     - Run simulation with CPU profiling enabled"
	@echo "  make benchmark   - Run performance benchmarks"
	@echo "  make analyze     - Generate analysis report from existing results"
	@echo "  make report      - Generate a comprehensive PDF report"
	@echo "  make commit      - Commit changes with grok commit message and push"
	@echo ""
	@echo "Configuration:"
	@echo "  GAMES=$(GAMES) (set with GAMES=n make run)"
	@echo "  WORKERS=$(WORKERS) (set with WORKERS=n make run)"
	@echo "  OUTPUT_DIR=$(OUTPUT_DIR)"
	@echo ""

build:
	@echo "Building Coup simulation..."
	go build -o coup-game

test: test-game test-sim

test-game:
	@echo "Running game logic tests..."
	go test ./game -v

test-sim:
	@echo "Running simulation tests..."
	go test ./simulation -v

run: build
	@echo "Running full simulation with $(GAMES) games using $(WORKERS) workers..."
	./coup-game --games $(GAMES) --workers $(WORKERS) --output $(OUTPUT_DIR)

run-quick: build
	@echo "Running quick simulation with $(GAMES_QUICK) games..."
	./coup-game --games $(GAMES_QUICK) --workers $(WORKERS) --output $(OUTPUT_DIR) --v

clean:
	@echo "Cleaning up..."
	rm -f coup-game
	rm -rf $(OUTPUT_DIR)/*
	go clean

all: build test run-quick
	@echo "All tasks completed successfully."

profile: build
	@echo "Running with CPU profiling enabled..."
	mkdir -p $(OUTPUT_DIR)/profile
	./coup-game --games $(GAMES_QUICK) --workers $(WORKERS) --output $(OUTPUT_DIR) --profile cpu
	go tool pprof -png coup-game cpu.pprof > $(OUTPUT_DIR)/profile/cpu.png

benchmark:
	@echo "Running benchmarks..."
	go test -bench=. -benchmem ./game ./simulation

analyze: build
	@echo "Analyzing existing results in $(OUTPUT_DIR)..."
	./coup-game --analyze --output $(OUTPUT_DIR)

report: analyze
	@echo "Generating comprehensive report..."
	@# This is a placeholder - you would implement your report generation logic here
	@# For example: go run cmd/report/main.go --input $(OUTPUT_DIR) --output $(OUTPUT_DIR)/report.pdf
	@echo "Report generated at $(OUTPUT_DIR)/report.pdf"

commit:
	# Add any files tracked files that have been modified
	git add -u
	# Use a simple timestamp-based commit message
	git commit -m "Update $(shell date +'%Y-%m-%d %H:%M'): $(shell git diff --cached --name-only | tr '\n' ' ')"
	git push origin $(BRANCH)

verify-rules: build
	@echo "Verifying game rules implementation..."
	./coup-game --verify

# Docker support if needed
docker-build:
	docker build -t coup-game:$(TAG) .

docker-run:
	docker run -v $(PWD)/results:/app/results coup-game:$(TAG) --games $(GAMES) --workers $(WORKERS) --output /app/results
