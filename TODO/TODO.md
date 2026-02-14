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

## Bugs (Open)

### BUG-8: Per-player-count win rates are fabricated, not measured
**File:** `simulation/metrics.go:256-282`

`GetStatisticsByPlayerCount()` doesn't track actual per-player-count character win
rates. Instead it applies `math.Cos(float64(playerCount))` as a "variance" factor to
the global win rate. This produces meaningless numbers in the player_count_analysis.csv.

**Fix:** Actually track character wins per player count during `ProcessGameResults`.

---

### BUG-9: Character `GamesPlayed` is estimated, not tracked
**File:** `simulation/metrics.go:93-98`

Instead of tracking which characters appeared in each game, `GamesPlayed` is set to
`totalGames * 3 / 5` for every character. This makes all character participation counts
identical and only approximate.

**Fix:** Track actual character participation per game using the winner's cards and/or
initial card deals (would require storing deal info in `GameResult`).

---

### BUG-10: Steal blocks double-count Captain and Ambassador
**File:** `simulation/metrics.go:170-176`

When a Steal is blocked, both Captain and Ambassador get credit for the block even
though only one character was actually claimed. This inflates block stats for both
characters.

**Fix:** Record the actual blocking character claimed in `ActionLog` and use that.

---

## Feature Requests

### ~~FEAT-1: Store simulation results in StatisticsResult for CSV export~~ FIXED (BUG-1)

### FEAT-2: Add proper statistical significance testing
**File:** `analysis/analyzer.go:177-204`

The current "significance" calculation uses hardcoded heuristic thresholds instead of
real hypothesis testing. Implement chi-squared or binomial proportion tests to
determine whether character win rate differences are statistically significant.

---

### FEAT-3: Track card information in GameResult
Currently `GameResult` only stores the winner's remaining cards. Tracking which cards
each player was dealt would enable accurate per-character statistics instead of the
current estimates.

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

### FEAT-5: Record actual blocking character in ActionLog
**File:** `game/game.go:22-30`

`ActionLog.Blocker` stores the blocker's player ID but not which character they claimed
to block with. Add a `BlockingCharacter string` field to enable accurate block
statistics.

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

### FEAT-8: Separate strategy creation from repetitive code
**File:** `game/ai_strategy.go`

The five `Create*Strategy` functions (Duke, Assassin, Captain, Ambassador, Contessa)
share ~80% identical structure. Refactor into a single factory function parameterized
by character name and a config table.

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

### FEAT-10: Add game replay / verbose single-game mode
Add a mode that runs a single game with detailed turn-by-turn output showing every
action, challenge, block, and card exchange. Useful for debugging and verifying game
logic.

---

### FEAT-11: Medium-competitive AI exchange is random, not strategic
**File:** `game/enhanced_player.go:611-643`

Medium and Low competitive AIs use a fully random card exchange for the Ambassador
ability. Medium AIs should have at least some preference for keeping useful cards
(e.g., Duke, Assassin) rather than pure randomness.

---

### ~~FEAT-12: `profile` flag is declared but never implemented~~ FIXED
Implemented full profiling support with `runtime/pprof`:
- `--profile=cpu` generates cpu.pprof for CPU profiling
- `--profile=mem` generates mem.pprof for memory profiling
Profiles can be analyzed with `go tool pprof`.

---

### FEAT-13: Survival time is never tracked
**File:** `simulation/metrics.go:53`

`CharacterStats.TotalSurvivalTurns` is initialized to 0 and never incremented. The
survival time stats in the CSV output are always 0.

**Fix:** Track the turn on which each player is eliminated during `processGameActions`.
