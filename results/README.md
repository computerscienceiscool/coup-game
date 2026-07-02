# Simulation Results — Provenance

All data in this directory was regenerated on **2026-07-02** with the fixed
engine (rules bugs BUG-11..17), the card-memory AI (FEAT-6), and the measured
metrics pipeline (BUG-18..22). Results produced before this date came from an
engine with serious rule bugs and a colliding seed scheme — do not mix the
two generations.

## Runs

| Location | Command | Games |
|---|---|---|
| `results/` (root) | `./coup-game --games 1000000 --workers 8 --seed 42 --ai mixed --output ./results` | 1,000,000 (200k each at 2–6 players) |
| `results/original/` | `./coup-game --games 200000 --workers 8 --seed 42 --ai original --output ./results/original` | 200,000 |
| `results/low/`, `medium/`, `high/`, `mixed/` | same, with `--ai <mode>` | 200,000 each |
| `../resultsofgame.md` | `./coup-game --test-comp 1000000 --seed 42` | 1,000,000 (4-player, mixed levels) |

Runs are reproducible: per-game seeds derive only from `--seed` and the game
ID (`game.MixSeed`), independent of worker scheduling.

## Files and metric definitions

**`character_stats.csv`** — one row per character:
- `DealtWinRate` — P(a player wins | they were dealt the character at game start). The primary strength metric.
- `FinalHandWinRate` — share of games whose winner *ended* the game holding the character (kept or acquired via Exchange).
- `ActionSuccessRate` — successes/attempts of the character's signature action (Tax, Steal, Assassinate, Exchange). Contessa has no signature action, so 0.
- `BlockSuccessRate` — blocks claiming this character that actually stopped the action. Every attempt counts, including blocks defeated by a challenge. Assassin blocks nothing, so 0.
- `BluffRate` — share of the character's claims (actions + blocks) made without holding it, from ground truth.
- `BluffSuccessRate` — bluffed claims that went unchallenged (a challenged bluff is always caught).
- `ChallengedRate` — share of the character's claims that drew a challenge.
- `AvgTurnsSurvived` — average turns survived by players dealt the character (a turn = one action taken; eliminated-seat skips are not counted).
- `PowerScore` — composite: `0.6*DealtWinRate + 0.15*ActionSuccessRate + 0.1*BlockSuccessRate + 0.1*BluffSuccessRate + 0.05*SurvivalRate`. A documented judgment call, not a fitted model.

**`game_logs.csv`** — one row per game: `GameID, PlayerCount, Winner, WinnerCharacters, TotalTurns, Date`. `TotalTurns` counts actions actually taken.

**`action_logs.csv`** — one row per action:
`GameID, Turn, Player, Action, Target, Success, Challenged, ActorHadCard, Blocker, BlockingCharacter, BlockChallenged, BlockSucceeded`.
`ActorHadCard`/`BlockerHadCard` are ground truth recorded by the engine for
bluff analysis; AI players never see them. `Blocker` records every block
attempt — check `BlockSucceeded` to see whether it stopped the action.

**`player_count_analysis.csv`** — `DealtWinRate` per character per player
count, plus the measured `AvgGameLength` (average actions per game) for that
player count.

## Known caveats

- ~8.8% of games (almost all short 2-player games) coincidentally play out
  identical action sequences. Their RNG streams are distinct (verified by
  `TestSeedingReproducibleAndDistinct`); with only a handful of actions per
  game, exact repeats are expected at this sample size.
- Dealt player-slots are not fully independent observations (two characters
  per hand, several slots per game), so the chi-squared p-value printed by
  the simulator is an approximation.
- Challenges and blocks are offered clockwise from the acting player,
  first-accepted-wins. Some ordering convention is unavoidable in a digital
  implementation, but it contributes to a measurable last-seat advantage —
  read per-seat statistics with that in mind.
