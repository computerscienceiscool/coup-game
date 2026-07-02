# Coup Game Simulation with Multi-Level AI

This enhanced version of the Coup game simulation incorporates the multi-level AI profiles (High, Medium, and Low competitive) for each character role in the game.

## Overview of Changes

The simulation has been updated to include:

1. **Enhanced AI Strategies**: Different competitive levels for each character role
2. **Character-Specific Behavior**: Each character now has unique decision-making patterns
3. **Customizable AI Types**: Ability to create games with specific AI player types
4. **Mixed AI Mode**: Ability to run simulations with a mix of different AI types
5. **Comparison Tools**: Utilities to compare performance across different AI levels

## AI Strategy System

The new AI strategy system is much more sophisticated than the original:

- **Competitive Levels**: Low, Medium, and High competitive settings for each character
- **Character Preferences**: AIs now have preferences for specific characters
- **Action Preferences**: AIs can prefer certain actions over others
- **Target Selection**: Intelligent selection of targets for actions
- **Character-Specific Bluff Rates**: Different bluffing probabilities for each character
- **Character-Specific Block Rates**: Different blocking probabilities for each character
- **Card Memory** (Medium/High): AIs count the face-up discard pile and each
  player's public claim history — impossible claims are always challenged,
  visibly-impossible bluffs are never made, and High AIs scale suspicion by
  claim history (see docs/specification.md)

## Character Profiles

Each character now has distinct behavior patterns at different competitive levels:

### Duke
- **High**: Aggressively takes Tax, blocks Foreign Aid often (70% bluff rate)
- **Medium**: More balanced Tax usage, sometimes blocks Foreign Aid (40% bluff rate)
- **Low**: Only uses Tax when holding Duke, rarely bluffs (10% bluff rate)

### Assassin
- **High**: Focuses on assassinations, targets threats (60% bluff rate)
- **Medium**: Uses assassination opportunistically (35% bluff rate)
- **Low**: Conservative assassination usage, random targets (15% bluff rate)

### Captain
- **High**: Frequently steals from rich players (65% bluff rate)
- **Medium**: Steals when advantageous (40% bluff rate)
- **Low**: Only steals when holding Captain (10% bluff rate)

### Ambassador
- **High**: Strategic card exchanges for specific characters (50% bluff rate)
- **Medium**: Occasional exchanges when needed (30% bluff rate)
- **Low**: Rarely uses exchange, even when holding Ambassador (5% bluff rate)

### Contessa
- **High**: Defensive master, almost always claims Contessa (80% bluff rate)
- **Medium**: Balanced defense, often claims Contessa (50% bluff rate)
- **Low**: Inconsistent defense, only claims Contessa when holding (15% bluff rate)

## Using the Enhanced Simulation

### Command-Line Options

New command-line options have been added:

```
--ai <mode>      AI mode: original, mixed, high, medium, low
--balance        Enable character balancing (equal distribution)
```

### Run with Different AI Types

```bash
# Original AI behavior
./coup-game --ai original

# All AIs at high competitive level
./coup-game --ai high

# All AIs at medium competitive level
./coup-game --ai medium

# All AIs at low competitive level
./coup-game --ai low

# Mixed AI levels and types
./coup-game --ai mixed
```

### Using the Makefile

```bash
# Build with enhanced AI
make build

# Run a quick test with a specific AI mode
make test AI_MODE=high

# Run simulation with specific AI types
make run-original
make run-low
make run-medium
make run-high
make run-mixed

# Run all AI types for comparison
make run-all GAMES=10000

# Run the demonstration script
make demo
```

## Example Results

The different AI types produce different results. For example, in high-competitive mode, characters like Duke and Captain have significantly higher win rates due to their aggressive strategies.

## Future Enhancements

Potential future enhancements include:

1. Machine learning-based AIs that adapt strategies based on game state
2. Tournament mode to pit different AI types against each other
3. Visual replay of interesting games
4. More sophisticated bluffing strategies based on risk assessment

## Credits

This enhanced AI system is based on the original Coup game simulation, with additional character-specific competitive profiles.
