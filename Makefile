# Makefile for Coup Simulation with Multi-level AI

# Default values
GAMES := 1000  # Default small number for quick testing
WORKERS := $(shell nproc || echo 4)
OUTPUT_DIR := ./results
AI_MODE := mixed  # Default to mixed AI
LEVEL := medium  # Default competitive level for non-mixed modes
BALANCE := false # Character balance (equal distribution)

# Executable name
EXECUTABLE := coup-game

# Source files
SOURCES := $(wildcard *.go) $(wildcard game/*.go) $(wildcard simulation/*.go) $(wildcard analysis/*.go)

# Enhanced sources - our new files
ENHANCED_SOURCES := ai_strategy.go enhanced_player.go game_creation.go enhanced_simulator.go main_updated.go

# Build the game with enhanced AI support
build:
	@echo "Building Coup simulation with multi-level AI..."
	@echo "Copying enhanced AI files..."
	cp /home/claude/ai_strategy.go game/
	cp /home/claude/enhanced_player.go game/
	cp /home/claude/game_creation.go game/
	cp /home/claude/enhanced_simulator.go simulation/
	cp /home/claude/main_updated.go main.go
	go build -o $(EXECUTABLE)

# Run a quick test with the specified AI mode
test: build
	@echo "Running quick test with $(AI_MODE) AI..."
	./$(EXECUTABLE) --games 5 --workers 1 --ai $(AI_MODE) --output $(OUTPUT_DIR)/test_$(AI_MODE) --v

# Run a simulation with original AI
run-original: build
	@echo "Running simulation with original AI..."
	./$(EXECUTABLE) --games $(GAMES) --workers $(WORKERS) --ai original --output $(OUTPUT_DIR)/original

# Run a simulation with low competitive AI
run-low: build
	@echo "Running simulation with low competitive AI..."
	./$(EXECUTABLE) --games $(GAMES) --workers $(WORKERS) --ai low --output $(OUTPUT_DIR)/low

# Run a simulation with medium competitive AI
run-medium: build
	@echo "Running simulation with medium competitive AI..."
	./$(EXECUTABLE) --games $(GAMES) --workers $(WORKERS) --ai medium --output $(OUTPUT_DIR)/medium

# Run a simulation with high competitive AI
run-high: build
	@echo "Running simulation with high competitive AI..."
	./$(EXECUTABLE) --games $(GAMES) --workers $(WORKERS) --ai high --output $(OUTPUT_DIR)/high

# Run a simulation with mixed competitive AI
run-mixed: build
	@echo "Running simulation with mixed competitive AI..."
	./$(EXECUTABLE) --games $(GAMES) --workers $(WORKERS) --ai mixed --output $(OUTPUT_DIR)/mixed

# Run simulations with all AI modes for comparison
run-all: build
	@echo "Running simulations with all AI modes..."
	./$(EXECUTABLE) --games $(GAMES) --workers $(WORKERS) --ai original --output $(OUTPUT_DIR)/original
	./$(EXECUTABLE) --games $(GAMES) --workers $(WORKERS) --ai low --output $(OUTPUT_DIR)/low
	./$(EXECUTABLE) --games $(GAMES) --workers $(WORKERS) --ai medium --output $(OUTPUT_DIR)/medium
	./$(EXECUTABLE) --games $(GAMES) --workers $(WORKERS) --ai high --output $(OUTPUT_DIR)/high
	./$(EXECUTABLE) --games $(GAMES) --workers $(WORKERS) --ai mixed --output $(OUTPUT_DIR)/mixed

# Run the full demo
demo: build
	@echo "Running demo script..."
	chmod +x /home/claude/demo.sh
	/home/claude/demo.sh

# Clean up build artifacts and results
clean:
	@echo "Cleaning up..."
	rm -f $(EXECUTABLE)
	rm -rf $(OUTPUT_DIR)/*

# Run a full benchmark with 100,000 games
benchmark: build
	@echo "Running benchmark with 100,000 games..."
	./$(EXECUTABLE) --games 100000 --workers $(WORKERS) --ai mixed --output $(OUTPUT_DIR)/benchmark

# Default target shows help
help:
	@echo "Coup Game Simulation with Multi-level AI - Available commands:"
	@echo "  make build       - Build the Coup simulation executable with multi-level AI"
	@echo "  make test        - Run a quick test with the default AI mode"
	@echo "  make run-original - Run a simulation with the original AI behavior"
	@echo "  make run-low     - Run a simulation with low competitive AI"
	@echo "  make run-medium  - Run a simulation with medium competitive AI"
	@echo "  make run-high    - Run a simulation with high competitive AI"
	@echo "  make run-mixed   - Run a simulation with mixed competitive AI"
	@echo "  make run-all     - Run simulations with all AI modes for comparison"
	@echo "  make demo        - Run the demonstration script"
	@echo "  make clean       - Remove build artifacts and result files"
	@echo "  make benchmark   - Run a full benchmark with 100,000 games"
	@echo ""
	@echo "Configuration:"
	@echo "  GAMES=$(GAMES) (set with GAMES=n make run-...)"
	@echo "  WORKERS=$(WORKERS) (set with WORKERS=n make run-...)"
	@echo "  OUTPUT_DIR=$(OUTPUT_DIR) (set with OUTPUT_DIR=path make run-...)"
	@echo "  AI_MODE=$(AI_MODE) (set with AI_MODE=mode make test)"
	@echo ""

.PHONY: build test run-original run-low run-medium run-high run-mixed run-all demo clean benchmark help
