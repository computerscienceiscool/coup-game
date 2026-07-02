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

## Engine Rules Bugs (Fixed 2026-07-01)

Found by auditing the 1M-game results against the code (impossible winner hands,
non-target blocks, and mass-duplicate games were all visible in results/*.csv).

### ~~BUG-11: Winning a challenge granted an extra influence card (CRITICAL)~~ FIXED
`RevealCard` only marked the revealed card `Shown` and left it in the hand, while
`ResolveChallenge`/`ResolveBlockChallenge` added a *new copy* of the card to the deck and
dealt a replacement — every successfully defended challenge grew the defender's hand by one.
In the 1M-game dataset 48.9% of winners ended with 3–10 influence cards. `RevealCard` now
removes and returns the card, that same instance is shuffled back, and a replacement is
drawn, keeping hand size constant. The now-unused `Card.Shown` field was removed, along with
the "lose shown cards first" logic in `LoseInfluence`. `Game.ValidateInvariants()` enforces
the card economy (15 cards, 3 per character, ≤2 per hand) and is exercised after every
action in `TestCardConservation`.

### ~~BUG-12: Any player could block Steal/Assassinate (CRITICAL)~~ FIXED
`ResolveAction` offered the block to every living opponent and neither AI checked whether it
was the target — 40.3% of steal/assassination blocks in the dataset came from non-targets.
`potentialBlockers` now restricts blocking to the target for targeted actions (any player
may still block Foreign Aid), enforced at the engine level. Covered by `TestOnlyTargetCanBlock`.

### ~~BUG-13: Lost influence returned to the deck~~ FIXED
Lost cards were shuffled back into the deck (a house rule the spec had codified). Per
official rules they're now removed from play into `Game.Discarded`, which also enables
future card-counting AIs. The spec was corrected to match.

### ~~BUG-14: Per-game seeds collided across workers (CRITICAL for statistics)~~ FIXED
`gameSeed = Seed + workerID + gameID` collides whenever the sums match: 67.5% of the 1M
games were bit-identical duplicates of another game, and results weren't reproducible even
with a fixed seed (scheduling decided which worker got which game). Seeds are now derived
via `game.MixSeed` (SplitMix64) from the base seed and game ID only. Player RNGs were also
double-offset (`seed+i` at the call site plus `+id` in the constructor); they now get
independent streams via `MixSeed`. Covered by `TestSeedingReproducibleAndDistinct`.

### ~~BUG-15: Assassination cost not refunded on a successful challenge~~ FIXED
The 3 coins are paid up front and stay paid when blocked (correct), but the rulebook returns
the cost when the action itself is successfully challenged. `ResolveAction` now refunds it.
Covered by `TestAssassinateRefundOnSuccessfulChallenge` / `TestAssassinateCostPaidWhenBlocked`.

### ~~BUG-16: Turn counter counted skipped eliminated seats~~ FIXED
`NextPlayer` incremented `Turn` even when passing over dead players, inflating game-length
and survival statistics in proportion to the body count. Turns are now counted in
`ExecuteTurn` (one per action taken) and `NextPlayer` just advances to the next living seat.

### ~~BUG-17: An action could be blocked repeatedly / enhanced block logic early-returned~~ FIXED
After a block was defeated by a challenge, the loop let the next player also block the same
action; now an action is blocked at most once. `EnhancedAIPlayer.BlockDecision` also
returned early on the Captain block-rate roll without ever checking for a held Ambassador;
the decision now checks all real blocking characters before considering a bluff.

---

## Metrics & Analysis Bugs (Fixed 2026-07-02)

### ~~BUG-18: ActionSuccessRate is never populated~~ FIXED
Actions are now attributed to the character they claim (Tax→Duke, Steal→Captain, ...) and
tallied from the log; character_stats.csv reports real signature-action success rates.

### ~~BUG-19: Block statistics only count surviving blocks~~ FIXED
Defeated blocks stay in the ActionLog (`Blocker` is no longer reset); new `BlockChallenged`
and `BlockSucceeded` fields record the outcome. BlockSuccessRate = stopped/attempted.

### ~~BUG-20: "BluffSuccessRate" measures the opposite of its name~~ FIXED
The engine now records ground truth at claim time (`ActorHadCard`, `BlockerHadCard`), so
bluff metrics are direct measurements: BluffRate = claims made without the card,
BluffSuccessRate = bluffed claims that went unchallenged. Block claims are included, and
PowerScore no longer rewards being caught.

### ~~BUG-21: Per-player-count AvgGameLength is fabricated~~ FIXED
`GetStatisticsByPlayerCount` now reports the measured average (TurnsByPlayerCount /
GamesByPlayerCount). The `*(1+0.1*(pc-3))` formula is gone.

### ~~BUG-22: Character win attribution is a mismatched ratio~~ FIXED
Primary metric is now DealtWinRate = P(win | dealt the character at game start), computed
per player-slot; FinalHandWinRate (winner's final hand contained it, deduped) is reported
alongside. The chi-squared significance test now compares dealt wins against expectations
proportional to dealt counts (approximation caveat documented in analyzer.go).

### ~~FEAT-6: Support game state tracking for known cards~~ FIXED
GameState now carries real public information: the face-up discard pile and every player's
public claim history (cleared when they Exchange). EnhancedAIPlayer uses it: Medium and High
AIs auto-challenge claims whose three copies are all visible (own hand + discard) and never
make visibly-impossible bluffs; High AIs also scale their challenge rate by visible copies
and by how many distinct characters the claimant has claimed. Low AIs remain memoryless.
Effect measured over 1M games: skill levels went from Low 36.7% / High 30.9% of wins
(aggression punished) to Low 33.8% / High 33.5% (near parity).

---

## Bugs (Open)

None known. All results under results/ and resultsofgame.md were regenerated 2026-07-02
with the fixed engine and metrics (see results/README.md for provenance).

---

## Feature Requests

### ~~FEAT-1: Store simulation results in StatisticsResult for CSV export~~ FIXED (BUG-1)

### ~~FEAT-2: Add proper statistical significance testing~~ FIXED (with caveat)
`calculateSignificance` runs a chi-squared goodness-of-fit test of dealt wins against
expectations proportional to dealt counts. Dealt player-slots are not fully independent
observations (two characters per hand, several slots per game), so treat the p-value as
an approximation — documented in the code and results/README.md.

---

### ~~FEAT-3: Track card information in GameResult~~ FIXED
Added PlayerStartingCards map[int][]string field to GameResult to track which cards
each player was dealt at game start. This enables accurate per-character statistics.

---

### FEAT-4: Add comprehensive test coverage (mostly done)
Coverage is now 87.5% (game) and 79.9% (simulation): card-conservation invariants after
every action across all AI modes, target-only blocking, assassinate cost rules, defeated
block logging, claims tracking, card-counting challenges, metrics collection, and seeding
reproducibility/distinctness — all run clean under `-race`. Still missing: analysis/export
tests (0%) and Ambassador exchange edge cases.

---

### ~~FEAT-5: Record actual blocking character in ActionLog~~ FIXED
Added BlockingCharacter string field to ActionLog. The field is populated when a block
occurs and cleared if the block is successfully challenged.

---

### ~~FEAT-6: Support game state tracking for known cards~~ FIXED
See the entry under "Metrics & Analysis Bugs (Fixed 2026-07-02)": GameState now exposes
the discard pile and public claim history, and Medium/High AIs use them.

---

### FEAT-14: Randomize challenge/block priority
**File:** `game/game.go` (challenge and block loops)

Challenges and blocks are offered clockwise from the acting player, first-accepted-wins.
That convention is unavoidable in some form, but it concentrates challenge risk/reward on
the seats right after the actor and contributes to a measured last-seat advantage (21.2%
vs 14.2% fair-share-16.7% in 6-player games). Consider randomizing the polling order per
action, or collecting all willing challengers and picking one at random, to better model
the simultaneous "speak up" rule of tabletop play.

---

### FEAT-15: Deeper deduction AI
Card memory (FEAT-6) covers certainty: impossible claims and impossible bluffs. A full
deduction AI would track probabilities — Bayesian updating over opponents' likely hands
from their claims, blocks, and exchanges — and choose bluffs that are consistent with its
own claim history. This is the difference between counting cards and reading the table.

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
