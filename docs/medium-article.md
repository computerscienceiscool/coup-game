# Simulating One Million Games of Coup: What the Data Really Says About Bluffing, Strategy, and Character Balance

*How I built a high-performance game simulator in Go, found the bugs in my own million-game dataset, fixed them, taught the AIs to count cards, and got a different answer than I started with.*

---

## The Question That Started It All

If you've ever played Coup, the indie bluffing card game that's taken over game nights worldwide, you've probably had this argument: **which character is the best?**

Is it the Duke, with its reliable income-generating Tax action? The Assassin, who can eliminate opponents for just 3 coins? Or maybe the Ambassador, with its defensive flexibility and card-swapping ability?

After countless heated debates at my local game night, I decided to settle this scientifically. I built a simulator in Go, ran a million games, wrote up the results — and then discovered, by auditing my own data, that the simulator wasn't playing Coup by the rules. This article is the story of both halves: the answer, and what it took to make the answer trustworthy.

---

## What Is Coup?

For the uninitiated, Coup is a social deduction game where players claim to have character cards (Duke, Assassin, Captain, Ambassador, Contessa) to perform powerful actions. The catch? You can lie about which cards you have. Opponents can challenge your claims or block your actions, creating a tense psychological battle of bluffing and deduction.

The game has simple rules but deep strategy:
- **Tax (Duke)**: Take 3 coins
- **Assassinate (Assassin)**: Pay 3 coins to eliminate someone's influence
- **Steal (Captain)**: Take 2 coins from another player
- **Exchange (Ambassador)**: Swap your cards with new ones from the deck
- **Block Foreign Aid (Duke)**: Stop someone from taking 2 coins
- **Block Assassination (Contessa)**: Stop someone from assassinating you — only if you're the target
- **Block Stealing (Captain/Ambassador)**: Stop someone from stealing from you

Each player starts with 2 hidden character cards. When you lose both influences (by failed challenges or successful attacks), you're eliminated. Last player standing wins.

---

## Building the Simulator

The simulator is written in Go — three components:

**Game engine** (`game/`): the full action/challenge/block loop, clockwise resolution order, forced Coup at 10+ coins, lost influence face-up and out of play, and an action log that records *ground truth* — whether each claim was actually a bluff — for analysis (the AIs never see it).

**AI players** (`game/enhanced_player.go`): three competitive levels with character preferences, from Low (5–15% bluff, 30% challenge, no memory) through Medium (30–50% bluff, 50% challenge) to High (50–80% bluff, 60–70% challenge). Crucially, Medium and High AIs now **count cards**: the engine exposes the public discard pile and every player's claim history, so a card-counting AI always challenges a claim whose three copies it can fully account for, never makes a visibly-impossible bluff, and (at High) scales its suspicion by how many distinct characters you've claimed.

**Simulation engine** (`simulation/`): a goroutine worker pool that runs 9,000–27,000 games per second on my 8-core machine, with per-game seeds derived only from the base seed and game ID — so runs are exactly reproducible regardless of scheduling.

The headline dataset: **1,000,000 games in mixed-AI mode (seed 42), 200,000 each at 2–6 players, 10.4 million logged actions.** Every number below comes from CSVs committed in the repo.

---

## The Headline: It Depends on the Table Size

The primary metric is the cleanest question the data can answer: **if you're dealt a character at the start of the game, how often do you go on to win?** Across all table sizes (a random dealt slot wins 25.3% of the time as baseline):

```
🥇 Captain      29.26%  ██████████████████████████████████████████████████
🥈 Duke         28.50%  █████████████████████████████████████████████████
🥉 Ambassador   26.81%  ██████████████████████████████████████████████
4️⃣  Assassin    22.06%  ██████████████████████████████████████
5️⃣  Contessa    19.71%  ██████████████████████████████████
```

**Captain, not Duke.** But the aggregate hides the real finding — the best card depends on how many people are at the table:

| Players | Best card | Runner-up |
|---|---|---|
| 2 | **Duke** 55.9% | Captain 53.3% |
| 3 | **Duke** 38.3% | Captain 37.2% |
| 4 | **Captain** 29.3% | Duke 28.8% |
| 5 | **Captain** 24.5% | Duke 22.9% |
| 6 | **Captain** 21.3% | Duke 19.0% |

Duke rules heads-up, where Tax's guaranteed economy races fastest. Captain takes over as the table grows: more opponents means more coins in circulation to steal, more steal targets, and — since Captain also *blocks* stealing — more chances for the card to matter. Contessa is last at every single table size (11.6% in 6-player games, barely two-thirds of its fair share).

Here's the twist that justifies the whole metric exercise: if you instead ask "which character did winners *end* the game holding?", **Duke comes out first** (30.6% of winning final hands vs Captain's 27.1%). Winners keep and acquire Dukes. But being *dealt* a Captain wins more games. Which one is "the strongest character"? Depends on the question — which is exactly why an article should say which question it's answering.

### The Contessa Problem, Confirmed

Contessa's story survived every bug fix: last place at every table size. It does one thing, only when you're targeted, and assassinations are just 8.5% of all actions (succeeding 30.7% of the time). One genuinely funny wrinkle: Contessa has the *best* bluff success rate in the game — 41.7% of Contessa claims made without the card went unchallenged, the highest of any character — because challenging a Contessa block means betting an influence against a card people actually tend to keep. The card is weak; *claiming* it is oddly safe.

---

## The Bluffing Economy

Ground-truth logging (the engine records whether each claim was real) turns bluffing from vibes into numbers, across 10.4 million actions:

- **~70% of all character claims are bluffs.** These AIs choose actions with only weak regard for their actual cards — much like your one friend.
- **70.2% of the 5.0 million challenges caught a bluff.** In a world where most claims are lies, challenging is very good business.
- Foreign Aid drew a block attempt 71.9% of the time, and 69.6% of those blocks stuck — the two-coin action succeeds only half the time.
- **The Steal paradox persists**: Steal is the most popular action (25.8% of all actions) and the least successful (19.8%) — 69.9% of attempts get challenged, and survivors still face target blocks that stick 80.7% of the time.

| Action | Share | Bluffed | Challenged | Succeeded |
|---|---|---|---|---|
| Steal | 25.8% | 69.9% | 69.9% | 19.8% |
| Tax | 20.0% | 70.4% | 65.6% | 53.4% |
| Exchange | 18.2% | 67.9% | 63.5% | 56.2% |
| Income | 16.2% | — | — | 100% |
| Foreign Aid | 10.4% | — | — | 50.0% |
| Assassinate | 8.5% | 69.7% | 59.5% | 30.7% |
| Coup | 1.0% | — | — | 100% |

Games are quick: 5.9 actions on average for 2 players, scaling to 14.4 for 6. Half of all winners (52.3%) limp across the finish line with a single influence left.

---

## Turn Order Matters — More Than I Expected

Win rate by seat position, one million games:

- **4 players** (fair share 25%): seat 1 wins 22.1% … seat 4 wins **29.2%**
- **6 players** (fair share 16.7%): seat 1 wins 14.2% … seat 6 wins **21.2%**

The last seat beats the first by 7 percentage points in a 6-player game. That's enormous. Part of this is real Coup — early actors expose claims while everyone is still alive and dangerous. But part of it is a simulator convention I want to flag honestly: challenges are offered clockwise from the actor, first-taker-wins, which concentrates challenge risk and reward on the seats right after the actor. Tabletop Coup resolves "who speaks first" by racing humans, which no simulator reproduces exactly. Treat the direction (late seats good) as believable and the magnitude as variant-specific.

---

## Does Skill Help? Card Counting Changed the Answer

I ran a second experiment: one million 4-player games where each seat gets a randomly chosen Low, Medium, or High competitive AI (fair share ≈33.3% each).

Before the AIs had card memory, this experiment embarrassed the "skilled" bots: **Low won 36.7%, High only 30.9%.** High AIs bluffed and challenged constantly in a world where every failed bluff and wrong challenge costs an influence — aggression was just self-harm with extra steps.

With card counting, the gap nearly closes: **Low 33.8%, Medium 32.7%, High 33.5%.** The High AIs' aggression stops being a tax once their challenges are informed (impossible claims are free wins) and their bluffs are never self-evidently fake. One amusing constant: High-competitive winners end the game holding an Ambassador far more often than anything else (19.4% of their final card slots) — the optimized Exchange logic hoards its favorite cards.

The honest caveat stands: this measures *these heuristics*, not human skill. The AIs count certainties; they don't yet do the Bayesian "you've claimed three different characters in five turns" glare that a good human brings.

---

## Before and After: What Fixing the Engine Changed

The first version of this analysis ran on an engine with real bugs, which I found by auditing the published CSVs — the data itself was the whistleblower. What changed when they were fixed:

| Measurement | Buggy engine | Fixed engine |
|---|---|---|
| Winners ending with impossible 3–10 card hands | 48.9% of games | 0 (max is 2, enforced by tests) |
| Steal/Assassinate blocks by non-targets (illegal) | 40.3% of blocks | 0 |
| Bit-identical duplicate games (seed collisions) | 67.5% | 8.8%, all coincidental short games |
| Average game length | 13.1 actions | 10.4 actions |
| Steal success rate | 12.4% | 19.8% |
| Assassination success rate | 25.6% | 30.7% |
| "Best character" | Duke, by every (broken) metric | Duke ≤3 players, Captain ≥4 |

The worst bug was subtle and delicious: when a challenged player *proved* their claim, the engine gave them a replacement card **without removing the revealed one** — every successfully defended challenge minted a free extra life. One "winner" finished holding ten influence cards in a game whose legal maximum is two. Nearly half the dataset's winners had impossible hands, and nobody noticed until the hand sizes were tabulated.

The moral I keep relearning: **a simulator is a claim about a game's rules, and your own output data is how the claim gets audited.** The fix suite now asserts card conservation — exactly 15 cards, at most 3 per character, at most 2 per hand — after *every action* of a thousand test games.

---

## How Much Should You Trust These Numbers?

More than the last batch. Specifically:

- **The engine is rule-faithful and test-enforced** (card economy invariants, target-only blocks, assassination cost rules, challenge reveal flow), and runs reproduce exactly from a seed.
- **Every metric is measured, never estimated** — earlier versions of the pipeline fabricated per-player-count game lengths from a formula and published always-zero action stats; the schema now documents each column's definition (see `results/README.md`).
- With ~1.5 million dealt slots per character, the win-rate differences dwarf sampling noise (the chi-squared test's p-value underflows to zero, though dealt slots aren't fully independent, so read it as "overwhelming," not as exact inference).
- Remaining caveats: 8.8% of games (almost all 6-action 2-player games) coincidentally replay identical sequences; the challenge-order convention inflates the seat effect; and the AIs, while now card-counting, are still heuristics — not humans, and not yet full deduction engines.

---

## What This Suggests at the Table

1. **Value Captain more, especially at full tables.** The consensus "Duke is best" holds only at 2–3 players in this data.
2. **Challenge more than feels polite.** When claims are mostly opportunistic, challengers profit. Your friends bluff more than you think.
3. **Steal less, or expect to be blocked.** Most-attempted, least-successful, in every version of this simulation.
4. **Contessa is a trap** — but *claiming* Contessa is the safest bluff in the game. Read into that what you will.
5. **Fight for a late seat.**

---

## Technical Deep Dive

For the technically inclined:

```go
// Worker pool pattern
for i := 0; i < numWorkers; i++ {
    go func(workerID int) {
        for gameID := range gameChannel {
            result := runGame(gameID)
            resultChannel <- result
        }
    }(i)
}
```

Lessons that cost me:

1. **Seed arithmetic is not free entropy.** `seed + workerID + gameID` collides whenever the sums match; two-thirds of my original million games were byte-identical duplicates, and results changed run-to-run because the OS scheduler picked which worker got which game. Per-game seeds now come from a SplitMix64 mix of (seed, gameID) alone.
2. **Log what you'll want to disprove.** Because every action, challenge, block, and ground-truth "was it a bluff" bit goes to CSV, the impossible 10-card winner was *findable* instead of just suspectable.
3. **Metrics need tests too.** A stats pipeline that outputs a plausible 75% will not warn you its denominator is wrong.
4. **Invariant tests beat example tests for game engines.** "After every action: 15 cards, ≤3 per character, ≤2 per hand" caught more than every handwritten scenario combined.

---

## Try It Yourself

The simulator is open source: [github.com/computerscienceiscool/coup-game](https://github.com/computerscienceiscool/coup-game)

```bash
git clone https://github.com/computerscienceiscool/coup-game
cd coup-game

# The headline run (about two minutes on 8 cores)
./coup-game --games 1000000 --workers 8 --seed 42 --ai mixed

# Compare AI modes
./coup-game --games 200000 --ai high
./coup-game --games 200000 --ai original

# Skill-level experiment
./coup-game --test-comp 1000000 --seed 42

# Watch a single game play out
./coup-game --replay --seed 42 --ai mixed
```

Results export to CSV (schemas documented in `results/README.md`) for analysis in Excel, Python, or R.

---

## Future Work

1. **Deduction AI**: Bayesian tracking of opponents' likely hands from claims, blocks, and exchanges — counting cards is arithmetic; reading the table is inference
2. **Randomized challenge priority**, to separate the real last-seat advantage from the polling-order artifact
3. **House-rule experiments** (nerfed Duke, Contessa-blocks-Steal) — now that the baseline is trustworthy, balance patches can be measured against it
4. **Tournament mode** pitting AI strategies head-to-head
5. **Human vs. AI** CLI play mode

---

## Conclusion

I started this project to settle an argument, and the data did settle it — just not the way either side expected. *"Which character is strongest?"* turns out to be underspecified: **Duke, if it's date night; Captain, once the whole group shows up; never Contessa.** The deeper findings were about the game's texture: most claims are lies, challenging liars is profitable, the boring guaranteed actions are underrated, and the person who acts last has a real edge.

But the finding I'll actually carry forward is about simulation itself. My first million games produced confident, plausible, wrong numbers — and the same CSVs that spread the wrong answer contained everything needed to catch it. Simulate boldly; then interrogate your own output like it's your opponent's third Duke claim of the night.

---

## Technical Stats

- **Language**: Go 1.22
- **Lines of code**: ~5,500 (including tests)
- **Test coverage**: 87.5% (game engine), 79.9% (simulation), enforced invariants under `-race`
- **Performance**: 9,000–27,000 games/second on 8 cores, depending on AI mode
- **Games behind this article**: 3,000,000 (1M mixed + 1M skill experiment + 1M across five AI modes), all seed 42, all reproducible
- **Actions logged**: 10.4 million in the main run
- **Most humbling bug**: a "winner" holding **10 influence cards** in a game whose legal maximum is 2 — undetected across 1,000,000 games until the data was audited

---

*Want to discuss game balance, simulation techniques, or argue about whether Captain is really better than Duke? Check out the project on GitHub.*
