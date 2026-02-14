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

### DESIGN-1: Merge Simulator and EnhancedSimulator (code duplication)
**Files:** `simulation/simulation.go`, `simulation/enhanced_simulator.go`

`Simulator` and `EnhancedSimulator` share ~90% identical code (`Run`, `worker`,
`collectResults`, `submitGames`, `determinePlayerCount`, `updateProgress`). They differ
only in how the game is created inside `worker()`. The basic `Simulator` is never called
from `main.go`.

**Fix:** Delete `Simulator`. Extract game creation into a `GameFactory` interface:
```go
type GameFactory func(playerCount int, seed int64) (*game.Game, error)
```
Pass it to a single `Simulator` via config. `EnhancedSimulator.worker` already has a
switch on `AIMode` — that switch becomes the factory.

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

### DESIGN-3: Action log export is arbitrarily capped at 10k rows
**File:** `analysis/export.go:162-168`

`generateActionLogs` hard-caps output at `maxActions = 10000`. With 1M games averaging
~15 actions each, this captures <0.1% of data with no warning to the user.

**Fix:** Either:
- Remove the cap and write all actions (stream to disk, don't hold in memory)
- Make it configurable via a flag (`--max-action-logs`)
- Sample uniformly instead of taking only the first N

---

### DESIGN-4: `game.Verbose` is global mutable state
**File:** `game/game_creation.go` (uses `game.Verbose`), `main.go:90`

`game.Verbose` is a package-level `var` set from `main()`. This is not goroutine-safe
(concurrent workers could read it while main writes it) and makes testing harder.

**Fix:** Pass verbosity through the `Game` struct or the creation function parameters
instead of relying on a global.

---

### DESIGN-5: Challenge/block ordering bias (player ID 0 always goes first)
**File:** `game/game.go:204-227` (challenge loop), `game/game.go:233-286` (block loop)

Both loops iterate `g.Players` in index order starting from 0. This means Player 0
always gets the first opportunity to challenge or block. In real Coup, any player can
react and the order shouldn't systematically favor lower IDs across millions of games.

**Fix:** Start iteration from a random offset or from the player to the left of the
acting player (clockwise), which matches tabletop convention:
```go
for offset := 1; offset < len(g.Players); offset++ {
    i := (g.CurrentPlayer + offset) % len(g.Players)
    opponent := g.Players[i]
    // ...
}
```

---

### DESIGN-6: Duplicate stat types across `analysis` and `simulation`
**Files:** `analysis/analyzer.go`, `simulation/simulation.go`, `simulation/metrics.go`

Both packages define their own `CharacterRanking`, `CharacterStats`/`CharacterStatistics`,
and `PlayerCountStats`/`PlayerCountStatistics`. The `Analyzer` just copies fields from
one to the other in `analyzeCharacters()` and `convertRankings()`.

**Fix:** Use the `simulation` types directly in `analysis`, or define shared types in a
`model` package. Eliminates ~60 lines of boilerplate conversion code.

---

### DESIGN-7: Error handling in simulation workers swallows errors
**File:** `simulation/enhanced_simulator.go:132-134`

When game creation fails, the error is printed to stdout and the game is silently
skipped. The final results count will be less than `TotalGames` with no indication.

**Fix:** Track errors in the result channel or a separate error channel:
```go
type GameOutcome struct {
    Result GameResult
    Err    error
}
```
Report error count and affected game IDs at simulation end.

---

### DESIGN-8: CSV output has non-deterministic row order
**File:** `analysis/export.go:228-243`

`generatePlayerCountAnalysis` iterates `stats.PlayerCountStats` (a map) and nested
`CharacterWinRates` (also a map). Map iteration order in Go is randomized, so the CSV
rows appear in different order each run, making diffs impossible.

**Fix:** Sort player counts and character names before writing:
```go
counts := sortedKeys(stats.PlayerCountStats)
for _, count := range counts { ... }
```

---

### DESIGN-9: `NewGame()` still uses hardcoded rates instead of constants
**File:** `game/game.go:63-66`

Despite FEAT-9 extracting magic numbers into `constants.go`, `NewGame()` still creates
AI players with literal `0.3` and `0.5`:
```go
players[i] = NewAIPlayer(i, &AIStrategy{
    BluffRate:     0.3,
    ChallengeRate: 0.5,
```

**Fix:** Use `DefaultBluffRate` and `DefaultChallengeRate` from `constants.go`.

---

### DESIGN-10: `calculateSignificance()` returns fake p-values
**File:** `analysis/analyzer.go:177-204`

The function returns hardcoded values (0.01, 0.05, 0.10) based on heuristic thresholds
rather than performing any actual hypothesis test. The returned value is labeled
`SignificanceLevel` which implies statistical rigor that doesn't exist.

**Fix:** Implement a chi-squared goodness-of-fit test comparing observed character win
rates against a uniform distribution (null hypothesis: all characters win equally). Go
has no stdlib chi-squared, but the test is ~20 lines of math. Alternatively, rename the
field to `ConfidenceHeuristic` to avoid misrepresentation.

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

### DESIGN-12: Survival time only tracks winners, not losers
**File:** `simulation/metrics.go:103-107`

FEAT-13 was marked fixed but only tracks winner survival (full game length).
Characters that are eliminated early get 0 survival turns, which makes the metric
meaningless for comparing fragile vs. durable characters.

**Fix:** Add elimination tracking to the game. When a player loses their last influence,
record `(playerID, turn)` in the `Game` struct. Pass this as a new field in `GameResult`:
```go
type GameResult struct {
    // ...
    EliminationTurns map[int]int  // playerID -> turn eliminated
}
```
Then compute survival per character from starting cards + elimination turn.
