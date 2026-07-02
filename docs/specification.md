# COUP CARD GAME SIMULATION SPECIFICATION

## Project Goal
Build a Go-based simulation engine to prove which Coup characters are statistically strongest through millions of simulated games with random AI players.

## Core Game Implementation

### Characters (Base Game)
1. **Duke** - Take 3 coins (Tax), Block foreign aid
2. **Assassin** - Pay 3 coins to assassinate, remove one influence
3. **Captain** - Steal 2 coins from another player, Block stealing
4. **Ambassador** - Exchange cards with Court Deck, Block stealing  
5. **Contessa** - Block assassination

### General Actions (Available to all)
- **Income** - Take 1 coin
- **Foreign Aid** - Take 2 coins (blockable by Duke)
- **Coup** - Pay 7 coins, remove one influence (unblockable)

## AI Player Behavior

### Original AI (baseline, `--ai original`)
- **Decision Making**: Random selection from all legal moves
- **Bluffing Rate**: 30% chance to claim a character they don't have
- **Challenge Rate**: 50% chance to challenge any claim
- **Blocking**: Always attempt to block if holding appropriate character
- **Target Selection**: Random when action requires a target

### Enhanced AI (`--ai low|medium|high|mixed`)
Character-preference strategies with per-level bluff/challenge/block rates
(see game/constants.go), weighted action selection, and targeting heuristics
(richest / strongest / biggest threat).

**Card memory.** The game exposes only information a real player could know:
the face-up discard pile and each player's public claim history (a completed
Exchange clears that player's history, since their cards may have changed).
Memory use scales with competitive level:
- **Low**: no memory — probability-only decisions
- **Medium**: counts certainties — always challenges a claim whose three
  copies are all visible (own hand + discard pile), and never makes a
  visibly-impossible bluff
- **High**: all of the above, plus scales its challenge rate by how many
  copies of the claimed character it can account for and by how many distinct
  characters the claimant has publicly claimed

The engine additionally records ground truth in the action log
(`ActorHadCard`, `BlockerHadCard`) for analysis only; AI players never see it.

## Simulation Parameters
- **Total Games**: 1 million
- **Distribution**: 200,000 games each for 2, 3, 4, 5, and 6 players
- **Deck Composition**: 3 copies of each character (15 cards total)
- **Starting Conditions**: 2 influences, 2 coins per player

## Metrics to Track

All metrics are measured from game data — never estimated or fabricated.
A "turn" is one action actually taken (skipped eliminated seats don't count).

### Primary Metrics
1. **Dealt Win Rate** - P(a player wins | dealt the character at game start),
   computed per player-slot. The primary strength metric.
2. **Character Power Score** - Composite:
   `0.6*DealtWinRate + 0.15*ActionSuccessRate + 0.1*BlockSuccessRate + 0.1*BluffSuccessRate + 0.05*SurvivalRate`
   (weights are a documented judgment call, not a fitted model)

### Secondary Metrics
3. **Final-Hand Win Rate** - Share of games whose winner ended the game holding
   the character (kept or acquired via Exchange), deduped per hand
4. **Action Success Rate** - Signature-action successes/attempts per character
   (Tax→Duke, Steal→Captain, Assassinate→Assassin, Exchange→Ambassador;
   Contessa has none)
5. **Block Success Rate** - Blocks claiming the character that stopped the
   action; every attempt counts, including blocks defeated by a challenge
6. **Bluff Rate** - Share of the character's claims (actions + blocks) made
   without holding it, from the ground-truth log fields
7. **Bluff Success Rate** - Bluffed claims that went unchallenged (a
   challenged bluff is always caught)
8. **Challenged Rate** - Share of the character's claims that drew a challenge
9. **Survival Time** - Average turns survived by players dealt the character

### Game-Level Metrics
10. **Game Length** - Actions per game, measured per player count
11. **Winner Character Composition** - Which characters winners were holding
12. **Statistical Significance** - Chi-squared goodness-of-fit of dealt wins
    vs. dealt-proportional expectations (approximate: dealt slots are not
    fully independent observations)

## Output Requirements

### CSV Files
1. **character_stats.csv** - Aggregated statistics per character
   - Columns: Character, DealtWinRate, FinalHandWinRate, ActionSuccessRate,
     BlockSuccessRate, BluffRate, BluffSuccessRate, ChallengedRate,
     AvgTurnsSurvived, TimesDealt, WinsWhenDealt, PowerScore
2. **game_logs.csv** - Individual game results
   - Columns: GameID, PlayerCount, Winner, WinnerCharacters, TotalTurns, Date
3. **action_logs.csv** - Detailed action-by-action data, including ground
   truth for bluff analysis (never exposed to AI players)
   - Columns: GameID, Turn, Player, Action, Target, Success, Challenged,
     ActorHadCard, Blocker, BlockingCharacter, BlockChallenged, BlockSucceeded
4. **player_count_analysis.csv** - Metrics broken down by player count
   - Columns: PlayerCount, Character, DealtWinRate, AvgGameLength

### Real-time Output
- Progress bar showing games completed
- Running statistics updated every 10,000 games
- ETA for completion

### Final Report
- Summary statistics in console
- Top 3 strongest characters with justification
- Interesting findings (e.g., "Duke wins 40% more in 2-player games")
- Statistical significance of results

## Technical Requirements

### Code Structure
```
coup-game/
├── main.go                 # Entry point, CLI
├── specification.md        # This document
├── go.mod                  # Go module file
├── game/
│   ├── game.go            # Core game logic
│   ├── player.go          # Player/AI logic  
│   ├── character.go       # Character definitions
│   ├── actions.go         # Action implementations
│   ├── deck.go            # Deck management
│   └── rules.go           # Game rules and validation
├── simulation/
│   ├── simulator.go       # Simulation engine
│   ├── metrics.go         # Metric tracking
│   ├── ai.go              # AI decision logic
│   └── config.go          # Simulation configuration
├── analysis/
│   ├── statistics.go      # Statistical analysis
│   ├── export.go          # CSV generation
│   └── report.go          # Report generation
└── tests/
    ├── game_test.go       # Game logic tests
    ├── simulation_test.go # Simulation tests
    └── rules_test.go      # Rule validation tests
```

### Performance Targets
- Minimum 1,000 games/second
- Memory efficient (handle millions of games)
- Concurrent processing using goroutines
- Configurable random seed for reproducibility
- Progress tracking without significant performance impact

### Game Engine Requirements

#### Game State Management
- Track all player influences (face-down cards)
- Track coin counts
- Track deck state
- Maintain action history
- Support game state validation

#### Action Resolution
1. Player declares action
2. Opponents can challenge (if applicable)
3. If challenged:
   - Challenged player reveals card or loses influence
   - If has card: challenger loses influence
   - If doesn't have card: challenged player loses influence
4. The action can be blocked (if applicable): only the target may block
   Steal/Assassinate; any player may block Foreign Aid
5. Blocker can be challenged
6. Action resolves if not successfully challenged/blocked (an action can be
   blocked at most once; a block defeated by a challenge lets the action proceed)

#### Challenge Logic
- Any character action can be challenged
- Blocks can be challenged
- General actions (Income, Foreign Aid, Coup) cannot be challenged
- Must track who can challenge (all other active players)

#### Elimination Rules
- Player eliminated when both influences lost
- Lost influence cards are revealed and permanently removed from play (face
  up, per official rules) — they never return to the deck
- Game continues until one player remains

## Validation Requirements
- Verify legal moves only
- Ensure proper challenge/block resolution
- Validate coin counts (cannot go negative)
- Validate forced coup at 10+ coins
- Ensure proper shuffle after returning cards
- Unit tests for all game rules
- Integration tests for complete games

## Configuration Options
```go
type SimulationConfig struct {
    TotalGames       int
    GamesPerPlayerCount map[int]int  // e.g., {2: 200000, 3: 200000, ...}
    BluffRate        float64         // 0.30
    ChallengeRate    float64         // 0.50
    AlwaysBlock      bool            // true
    RandomSeed       int64           // for reproducibility
    OutputDir        string          // for CSV files
    ProgressInterval int             // games between progress updates
}
```

## Success Criteria
1. Simulation completes 1 million games without errors
2. Statistical significance in character strength rankings
3. Clear identification of strongest/weakest characters
4. Reproducible results with same random seed
5. Performance meets 1,000+ games/second target
6. Memory usage remains stable during long runs
7. All game rules correctly implemented and validated

## Future Enhancements (Post-MVP)
- Implement smarter AI strategies
- Add Reformation expansion characters
- Support for tournament-style competitions
- Machine learning for optimal play
- Interactive game replay viewer
- Web dashboard for results visualization
- Configurable house rules support

