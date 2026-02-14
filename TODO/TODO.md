# Coup Game Simulation - TODO

## Bugs (Fixed)

### ~~BUG-1: `simulationResults()` always returns nil (CRITICAL)~~ FIXED
Removed the nil stub. `GenerateCSVs` now receives `*SimulationResults` directly and
passes it to `generateGameLogs` and `generateActionLogs`.

### ~~BUG-2: Makefile references nonexistent `/home/claude/` paths~~ FIXED
Removed `cp` commands from `build` target. Fixed `demo` target to use `./demo.sh`.

### ~~BUG-3: Assassinate doesn't cost coins when blocked~~ FIXED
Moved 3-coin deduction to `ResolveAction` (paid upfront before challenge/block phase).
Removed duplicate deduction from `AssassinateAction.Execute`.

### ~~BUG-4: `determinePlayerCount` nondeterministic map iteration~~ FIXED
Both `Simulator` and `EnhancedSimulator` now sort map keys before iterating.

### ~~BUG-5: Game creation functions unconditionally print to stdout~~ FIXED
All `fmt.Printf` calls in `game_creation.go` are now guarded with `if Verbose`.

### ~~BUG-6: Unchecked type assertions can panic~~ FIXED
All three target selection functions now use comma-ok type assertions.

### ~~BUG-7: No infinite game protection~~ FIXED
Added `MaxTurns = 500` constant. `RunToCompletion` force-ends the game and declares
the player with most influence (ties broken by coins) the winner.

---

## Bugs (Fixed)

### ~~BUG-8: Per-player-count win rates are fabricated, not measured~~ FIXED
Added CharacterWinsByPlayerCount map to track actual wins per character per player count.
GetStatisticsByPlayerCount() now calculates real win rates instead of using math.Cos() fabrication.

---

### ~~BUG-9: Character `GamesPlayed` is estimated, not tracked~~ FIXED
Added PlayerStartingCards field to GameResult to track which characters each player started with.
GamesPlayed is now incremented based on actual character participation, not estimated.

---

### ~~BUG-10: Steal blocks double-count Captain and Ambassador~~ FIXED
Added BlockingCharacter field to ActionLog to record which character was actually claimed.
Steal blocks now credit only the actual blocking character (Captain or Ambassador), not both.

---

## Bugs (Open)

None remaining!

---

## Feature Requests

### ~~FEAT-1: Store simulation results in StatisticsResult for CSV export~~ FIXED (BUG-1)

### FEAT-2: Add proper statistical significance testing
**File:** `analysis/analyzer.go:177-204`

The current "significance" calculation uses hardcoded heuristic thresholds instead of
real hypothesis testing. Implement chi-squared or binomial proportion tests to
determine whether character win rate differences are statistically significant.

---

### ~~FEAT-3: Track card information in GameResult~~ FIXED
Added PlayerStartingCards map[int][]string field to GameResult to track which cards
each player was dealt at game start. This enables accurate per-character statistics.

---

### FEAT-4: Add comprehensive test coverage
Current test coverage is minimal (4 game tests, 2 simulation tests, 0 analysis tests).
Missing tests:
- Challenge resolution (both success and failure)
- Block challenge mechanics
- Ambassador exchange edge cases (empty/near-empty deck)
- Enhanced AI strategy behavior at each competitive level
- Concurrent simulation correctness (race condition tests with `-race`)
- Analysis/export correctness

---

### ~~FEAT-5: Record actual blocking character in ActionLog~~ FIXED
Added BlockingCharacter string field to ActionLog. The field is populated when a block
occurs and cleared if the block is successfully challenged.

---

### FEAT-6: Support game state tracking for known cards
**File:** `game/game.go:389-414`

`GetStateForPlayer` returns empty `KnownCards` and `RemainingCards`. In real Coup,
players should track revealed cards from challenges. Implementing this would enable
smarter AI decision-making (e.g., not bluffing a character that's been revealed).

---

### ~~FEAT-7: Add a `--quiet` flag to suppress game-creation output~~ FIXED
Added `--quiet` flag that suppresses game-creation output, AI mode messages, test game
details, and intermediate status messages. Only displays the progress bar and final summary.

---

### ~~FEAT-8: Separate strategy creation from repetitive code~~ FIXED
Created strategyConfigs map with CharacterStrategyConfig for all character/level combinations.
Replaced five repetitive Create*Strategy functions with single createCharacterStrategy factory.
Reduced code duplication from ~200 lines to ~100 lines, improved maintainability.

---

### ~~FEAT-9: Extract magic numbers into named constants~~ FIXED
Created `game/constants.go` with all magic numbers extracted into named constants:
- Character-specific bluff/challenge rates for all competitive levels
- Action preference weights (Tax, Assassinate, Steal, Exchange, Income)
- Character-specific block rates
- Threat score multiplier (ThreatInfluenceMultiplier)
- Default/Original AI rates
All strategy functions and threat calculations now use these constants.

---

### ~~FEAT-10: Add game replay / verbose single-game mode~~ FIXED
Added --replay flag that runs a single game with detailed turn-by-turn output.
Shows initial game state, player strategies, every action with challenges/blocks,
and final results. Useful for debugging game logic.

---

### ~~FEAT-11: Medium-competitive AI exchange is random, not strategic~~ FIXED
Medium competitive AIs now use strategic card exchange with simplified scoring:
- Prefers Duke (1.5) and Assassin (1.2) over other characters
- Captain (0.8), Contessa (0.7), Ambassador (0.5) have lower priorities
- Adds 0.8 random variance (more than high competitive's 0.2)
Low competitive AI remains fully random for contrast.

---

### ~~FEAT-12: `profile` flag is declared but never implemented~~ FIXED
Implemented full profiling support with `runtime/pprof`:
- `--profile=cpu` generates cpu.pprof for CPU profiling
- `--profile=mem` generates mem.pprof for memory profiling
Profiles can be analyzed with `go tool pprof`.

---

### ~~FEAT-13: Survival time is never tracked~~ FIXED
TotalSurvivalTurns now tracks survival time for winner's characters (they survive the full game).
Winner characters accumulate result.TotalTurns for each game they win.

---

## Design Evaluation & Architectural Suggestions

### ~~DESIGN-1: Merge Simulator and EnhancedSimulator (code duplication)~~ FIXED
**Files:** `simulation/simulation.go`, ~~`simulation/enhanced_simulator.go`~~ (deleted)

Added AIMode, CompetitiveLevel, CharacterBalance fields to Config. Updated Simulator.worker()
to handle all AI modes (original, mixed, level-based). Added generateBalancedAITypes() method
to Simulator. Deleted enhanced_simulator.go (merged into simulation.go). Reduced code
duplication and simplified the codebase.

---

### DESIGN-2: Streaming metrics to reduce memory usage
**Files:** `simulation/enhanced_simulator.go`, `simulation/metrics.go`

Currently every `GameResult` (including its full `[]ActionLog`) is appended to
`s.Results` and held in memory until all games finish. At 1M games this is hundreds of
megabytes. The results slice is only needed so that `MetricsCollector.ProcessGameResults`
and `analysis.Analyzer` can iterate it after the fact.

**Fix:** Process metrics incrementally in `collectResults`:
```go
func (s *Simulator) collectResults(done chan bool) {
    for result := range s.ResultChannel {
        s.Metrics.ProcessGame(result)    // incremental
        // only store summary, not full ActionLog
    }
    done <- true
}
```
CSV game_logs can be written incrementally too (open file, append rows, close at end).
This drops memory from O(games * actions) to O(1).

---

### ~~DESIGN-3: Action log export is arbitrarily capped at 10k rows~~ FIXED
**File:** `analysis/export.go`

Removed the arbitrary `maxActions = 10000` cap. generateActionLogs() now writes all action
logs to CSV without limit, allowing full data export for large simulations.

---

### ~~DESIGN-4: `game.Verbose` is global mutable state~~ FIXED
**File:** `game/game_creation.go`, `main.go`

Removed the global `var Verbose` from game/game_creation.go. Removed all Verbose print
guards (debug output already handled by main.go's runTestGame). Removed unused helper
functions (getLevelName, getCharacterName, getCharacterPreferenceName). Eliminated
goroutine-unsafe global mutable state.

---

### ~~DESIGN-5: Challenge/block ordering bias (player ID 0 always goes first)~~ FIXED
**File:** `game/game.go` (challenge and block loops)

Implemented clockwise iteration starting from the acting player using offset pattern:
```go
for offset := 1; offset < numPlayers; offset++ {
    i := (g.CurrentPlayer + offset) % numPlayers
    opponent := g.Players[i]
    // ...
}
```
Both challenge and block loops now iterate clockwise from the acting player, matching
tabletop convention and eliminating systematic bias toward lower player IDs.

---

### DESIGN-6: Duplicate stat types across `analysis` and `simulation`
**Files:** `analysis/analyzer.go`, `simulation/simulation.go`, `simulation/metrics.go`

Both packages define their own `CharacterRanking`, `CharacterStats`/`CharacterStatistics`,
and `PlayerCountStats`/`PlayerCountStatistics`. The `Analyzer` just copies fields from
one to the other in `analyzeCharacters()` and `convertRankings()`.

**Fix:** Use the `simulation` types directly in `analysis`, or define shared types in a
`model` package. Eliminates ~60 lines of boilerplate conversion code.

---

### ~~DESIGN-7: Error handling in simulation workers swallows errors~~ FIXED
**File:** `simulation/simulation.go`

Added ErrorCount field to Simulator (int64, atomically incremented on failures).
Worker now uses `atomic.AddInt64(&s.ErrorCount, 1)` when game creation fails.
SimulationResults.ErrorCount is set from atomic load. main.go reports error count
if non-zero. Errors are now tracked and reported to the user.

---

### ~~DESIGN-8: CSV output has non-deterministic row order~~ FIXED
**File:** `analysis/export.go`

Added sort.Ints() for player counts and sort.Strings() for character names before
iterating maps. generatePlayerCountAnalysis() now produces deterministic CSV output,
making diffs and version control meaningful.

---

### ~~DESIGN-9: `NewGame()` still uses hardcoded rates instead of constants~~ FIXED
**File:** `game/game.go`

Updated NewGame() to use DefaultBluffRate and DefaultChallengeRate constants from
constants.go instead of hardcoded 0.3 and 0.5 values. All magic numbers now eliminated.

---

### ~~DESIGN-10: `calculateSignificance()` returns fake p-values~~ FIXED
**File:** `analysis/analyzer.go`

Implemented proper chi-squared goodness-of-fit test with Wilson-Hilferty approximation
for p-value calculation. Tests null hypothesis that all characters win at equal rates.
Returns actual p-value (0-1) instead of hardcoded thresholds. Provides real statistical
rigor for significance testing.

---

### DESIGN-11: No separation between game rules and AI decision-making
**Files:** `game/game.go`, `game/player.go`, `game/enhanced_player.go`

The `Player` interface mixes game-mechanical operations (`AddCoins`, `LoseInfluence`,
`HasCard`) with AI decisions (`ChooseAction`, `ChallengeDecision`, `BlockDecision`).
This makes it impossible to test game rules independently of AI behavior.

**Fix:** Split into two interfaces:
```go
type PlayerState interface {
    GetID() int
    GetCoins() int
    IsAlive() bool
    // ... state queries and mutations
}

type PlayerStrategy interface {
    ChooseAction(state GameState, actions []Action) Action
    ChallengeDecision(state GameState, actor Player, action Action) bool
    BlockDecision(state GameState, actor Player, action Action) bool
    // ... decisions
}
```
This also enables swapping strategies at runtime or testing rules with mock strategies.

---

### ~~DESIGN-12: Survival time only tracks winners, not losers~~ FIXED
**Files:** `game/game.go`, `simulation/simulation.go`, `simulation/metrics.go`

Added EliminationTurns map[int]int to Game struct to track when each player is eliminated.
CheckWinCondition() records elimination turn for dead players. GameResult includes
EliminationTurns field. Metrics now compute accurate per-character survival from starting
cards + elimination turn data. All three game creation functions initialize the map.
