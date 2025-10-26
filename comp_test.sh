#!/bin/bash
# comp_test.sh - Test win rates by competitiveness level

# Number of games to run
GAMES=100  # Starting with 100 games for faster testing

# Create results directory
mkdir -p results/comp_test

echo "Running competitive level analysis with $GAMES games..."

# Run games and capture output
./coup-game --games $GAMES --ai mixed -v > results/comp_test/game_output.txt

# Analyze the results using simple grep and counting
echo "Analyzing results..."

# Count total games with winners
TOTAL=$(grep -c "Winner strategy:" results/comp_test/game_output.txt)
echo "Total games analyzed: $TOTAL"

# Count wins by competitive level
LOW_WINS=$(grep -c "Winner strategy: Low Competitive" results/comp_test/game_output.txt)
MEDIUM_WINS=$(grep -c "Winner strategy: Medium Competitive" results/comp_test/game_output.txt)
HIGH_WINS=$(grep -c "Winner strategy: High Competitive" results/comp_test/game_output.txt)
ORIGINAL_WINS=$(grep -c "Winner strategy: Original AI" results/comp_test/game_output.txt)

# Calculate percentages
LOW_PERCENT=$(echo "scale=2; $LOW_WINS * 100 / $TOTAL" | bc)
MEDIUM_PERCENT=$(echo "scale=2; $MEDIUM_WINS * 100 / $TOTAL" | bc)
HIGH_PERCENT=$(echo "scale=2; $HIGH_WINS * 100 / $TOTAL" | bc)
ORIGINAL_PERCENT=$(echo "scale=2; $ORIGINAL_WINS * 100 / $TOTAL" | bc)

# Print results
echo ""
echo "COMPETITIVE LEVEL WIN RATES"
echo "==========================="
echo "Low Competitive:    $LOW_WINS wins ($LOW_PERCENT%)"
echo "Medium Competitive: $MEDIUM_WINS wins ($MEDIUM_PERCENT%)"
echo "High Competitive:   $HIGH_WINS wins ($HIGH_PERCENT%)"
echo "Original AI:        $ORIGINAL_WINS wins ($ORIGINAL_PERCENT%)"
echo "Total Games:        $TOTAL"

# Character analysis by level
echo ""
echo "CHARACTER WIN RATES (All Levels)"
echo "==============================="
echo "Duke:       $(grep -c "Winner's cards: \[Duke" results/comp_test/game_output.txt) wins"
echo "Captain:    $(grep -c "Winner's cards: \[Captain" results/comp_test/game_output.txt) wins"
echo "Assassin:   $(grep -c "Winner's cards: \[Assassin" results/comp_test/game_output.txt) wins"
echo "Ambassador: $(grep -c "Winner's cards: \[Ambassador" results/comp_test/game_output.txt) wins"
echo "Contessa:   $(grep -c "Winner's cards: \[Contessa" results/comp_test/game_output.txt) wins"

# Analyze winning cards by competitive level
echo ""
echo "TOP WINNING COMBINATIONS BY COMPETITIVE LEVEL"
echo "============================================"

echo "Low Competitive Winners:"
grep -B 2 "Winner strategy: Low Competitive" results/comp_test/game_output.txt | grep "Winner's cards:" | sort | uniq -c | sort -nr | head -5

echo "Medium Competitive Winners:"
grep -B 2 "Winner strategy: Medium Competitive" results/comp_test/game_output.txt | grep "Winner's cards:" | sort | uniq -c | sort -nr | head -5

echo "High Competitive Winners:"
grep -B 2 "Winner strategy: High Competitive" results/comp_test/game_output.txt | grep "Winner's cards:" | sort | uniq -c | sort -nr | head -5

echo ""
echo "Analysis complete! Full results saved to results/comp_test/game_output.txt"
