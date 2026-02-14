#!/bin/bash
# demo.sh - Demonstration of multi-level AI strategies in Coup

# Ensure the output directory exists
mkdir -p results

# Build the game
echo "Building Coup simulation..."
go build -o coup-game

# Function to print a section header
print_header() {
    echo ""
    echo "==================================="
    echo "$1"
    echo "==================================="
    echo ""
}

# Run test games with different AI levels
print_header "Test Game with Original AI"
./coup-game --games 1 --workers 1 --ai original -v

print_header "Test Game with Low Competitive AI"
./coup-game --games 1 --workers 1 --ai low -v

print_header "Test Game with Medium Competitive AI"
./coup-game --games 1 --workers 1 --ai medium -v

print_header "Test Game with High Competitive AI"
./coup-game --games 1 --workers 1 --ai high -v

print_header "Test Game with Mixed Competitive AI"
./coup-game --games 1 --workers 1 --ai mixed -v

# Run small simulations to compare results
print_header "Running Quick Simulation Comparisons"

echo "Running 100 games with Original AI..."
./coup-game --games 100 --workers 4 --ai original --output results/original

echo "Running 100 games with Low Competitive AI..."
./coup-game --games 100 --workers 4 --ai low --output results/low

echo "Running 100 games with Medium Competitive AI..."
./coup-game --games 100 --workers 4 --ai medium --output results/medium

echo "Running 100 games with High Competitive AI..."
./coup-game --games 100 --workers 4 --ai high --output results/high

echo "Running 100 games with Mixed Competitive AI..."
./coup-game --games 100 --workers 4 --ai mixed --output results/mixed

# Compare win rates
print_header "Comparing Character Win Rates"
echo "Original AI:"
cat results/original/character_stats.csv | head -n 6

echo "Low Competitive AI:"
cat results/low/character_stats.csv | head -n 6

echo "Medium Competitive AI:"
cat results/medium/character_stats.csv | head -n 6

echo "High Competitive AI:"
cat results/high/character_stats.csv | head -n 6

echo "Mixed Competitive AI:"
cat results/mixed/character_stats.csv | head -n 6

print_header "Demonstration Complete"
echo "Full results available in the results directory"
