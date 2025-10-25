# COUP SIMULATION DEVELOPMENT GUIDE

## Development Priority Order
Build incrementally, testing each phase before moving forward:

### Phase 1: Core Game Engine (Week 1)
1. **Game State Management**
   - Define all structs (Game, Player, Card, Deck)
   - Implement deck shuffle and deal
   - Track coins and influences

2. **Basic Actions**
   - Implement Income, Foreign Aid, Coup
   - Add validation for each action

3. **Character Actions**
   - Implement character-specific actions
   - Add blocking mechanics
   - Add challenge mechanics

4. **Game Flow**
   - Turn management
   - Elimination detection
   - Win condition checking

### Phase 2: AI Implementation (Week 1-2)
1. **Random AI**
   - Implement basic Player interface
   - Random action selection from legal moves
   - 30% bluff rate, 50% challenge rate

2. **Target Selection**
   - Random target selection
   - Validation of legal targets

### Phase 3: Simulation Framework (Week 2)
1. **Single Game Runner**
   - Complete game execution
   - Basic logging

2. **Batch Simulator**
   - Run multiple games
   - Basic metrics collection

3. **Concurrent Processing**
   - Worker pool pattern
   - Channel-based communication

### Phase 4: Metrics & Analysis (Week 2-3)
1. **Metric Collection**
   - Implement all metrics from spec
   - In-memory aggregation

2. **Statistical Analysis**
   - Win rate calculations
   - Character strength scoring

3. **CSV Export**
   - All required output files
   - Progress reporting

### Phase 5: Optimization (Week 3)
1. **Performance Profiling**
2. **Memory Optimization**
3. **Concurrency Tuning**

## Go-Specific Implementation Guidelines

### Code Style
```go
// GOOD: Idiomatic Go
func (g *Game) ExecuteAction(action Action) error {
    if err := g.validateAction(action); err != nil {
        return fmt.Errorf("invalid action: %w", err)
    }
    return action.Execute(g)
}

// AVOID: Too clever
func (g *Game) ExecuteAction(a Action) error {
    return func() error {
        if e := g.validateAction(a); e != nil { return e }
        return a.Execute(g)
    }()
}
```

### Error Handling
```go
// Always wrap errors with context
if err := player.TakeAction(); err != nil {
    return fmt.Errorf("player %d action failed: %w", player.ID, err)
}

// Use custom error types for game logic
type IllegalActionError struct {
    Player int
    Action string
    Reason string
}

func (e IllegalActionError) Error() string {
    return fmt.Sprintf("player %d cannot %s: %s", e.Player, e.Action, e.Reason)
}
```

### Concurrency Pattern
```go
// Worker pool for simulations
func RunSimulations(config Config) {
    games := make(chan GameConfig, 100)
    results := make(chan GameResult, 100)
    
    // Start workers
    var wg sync.WaitGroup
    for i := 0; i < runtime.NumCPU(); i++ {
        wg.Add(1)
        go worker(games, results, &wg)
    }
    
    // Feed games
    go func() {
        for i := 0; i < config.TotalGames; i++ {
            games <- NewGameConfig(i)
        }
        close(games)
    }()
    
    // Collect results
    go func() {
        wg.Wait()
        close(results)
    }()
    
    // Process results
    for result := range results {
        processResult(result)
    }
}
```

## Key Interfaces

### Core Interfaces
```go
package game

// Player represents any player (human or AI)
type Player interface {
    // Decision making
    ChooseAction(state GameState) Action
    ChallengeDecision(claim Claim, claimant *Player) bool
    BlockDecision(action Action, actor *Player) bool
    
    // When challenged
    RevealCard(cards []Card) Card  // Choose which card to reveal
    ChooseExchange(hand []Card, deck []Card) []Card  // Ambassador exchange
    
    // Information
    GetID() int
    GetCoins() int
    GetInfluences() []Card
    IsAlive() bool
}

// Action represents any game action
type Action interface {
    Execute(g *Game) error
    IsLegal(g *Game, p Player) error
    RequiresTarget() bool
    CanBeBlocked() bool
    CanBeChallenged() bool
    GetRequiredCard() Card
}

// Metric collector interface
type MetricCollector interface {
    RecordAction(action Action, success bool)
    RecordChallenge(challenger, challenged Player, success bool)
    RecordBlock(blocker Player, action Action, success bool)
    RecordElimination(player Player, turn int)
    RecordGameEnd(winner Player, turns int)
    GetStatistics() Statistics
}
```

### AI Strategy Interface (for future expansion)
```go
type Strategy interface {
    CalculateBluffProbability(state GameState) float64
    CalculateChallengeProbability(claim Claim, state GameState) float64
    SelectAction(legalActions []Action, state GameState) Action
    SelectTarget(validTargets []Player, action Action) Player
}
```

## Testing Requirements

### Test Coverage Goals
- **Game Logic**: 100% coverage (critical)
- **AI Logic**: 80% coverage
- **Simulation**: 70% coverage
- **Analysis**: 60% coverage

### Essential Test Cases
```go
// Test file structure
game/
├── game_test.go           // Core game mechanics
├── actions_test.go        // Each action type
├── challenge_test.go      // Challenge resolution
├── elimination_test.go    // Edge cases in elimination
└── integration_test.go    // Full game scenarios

// Example test
func TestForcedCoup(t *testing.T) {
    g := NewGame()
    p := g.Players[0]
    p.Coins = 10
    
    actions := g.GetLegalActions(p)
    
    // Should only have Coup as option
    assert.Len(t, actions, 1)
    assert.IsType(t, &CoupAction{}, actions[0])
}
```

### Edge Cases to Test
```go
// Must handle these scenarios:
- "Player has 10+ coins, must Coup, but all opponents have 1 influence"
- "Ambassador exchange with 0 or 1 cards left in deck"  
- "Captain steals from player with 1 coin"
- "Two players eliminate each other simultaneously"
- "Assassinate with exactly 3 coins"
- "Challenge a block that succeeds"
- "Deck runs out during game"
- "All players except one eliminated in same round"
```

### Benchmark Tests
```go
func BenchmarkFullGame(b *testing.B) {
    for i := 0; i < b.N; i++ {
        g := NewGame(4) // 4 players
        g.RunToCompletion()
    }
}

func BenchmarkShuffle(b *testing.B) {
    d := NewDeck()
    for i := 0; i < b.N; i++ {
        d.Shuffle()
    }
}
```

## Performance Optimization Guidelines

### Pre-Optimization Checklist
1. ✅ Working implementation with tests
2. ✅ Benchmark baseline established  
3. ✅ Profile to identify bottlenecks
4. ✅ Optimize only the hot paths

### Memory Optimization
```go
// Pre-allocate slices
players := make([]Player, 0, maxPlayers)

// Reuse objects with sync.Pool
var gameStatePool = sync.Pool{
    New: func() interface{} {
        return &GameState{
            Players: make([]PlayerState, 0, 6),
        }
    },
}

// Use value receivers for small structs
func (c Card) String() string { // Card is small, use value
    return c.Name
}

func (g *Game) Start() { // Game is large, use pointer
    // ...
}
```

### Concurrency Optimization
```go
// Buffer channels appropriately
results := make(chan Result, runtime.NumCPU()*2)

// Use atomic operations for counters
var gamesCompleted uint64
atomic.AddUint64(&gamesCompleted, 1)

// Minimize lock contention
type SafeMetrics struct {
    mu sync.RWMutex
    data map[string]int
}

func (m *SafeMetrics) Get(key string) int {
    m.mu.RLock() // Read lock for getters
    defer m.mu.RUnlock()
    return m.data[key]
}
```

## Debugging Features

### Verbose Mode
```go
type Config struct {
    Verbose      bool
    LogActions   bool
    LogChallenges bool
    SaveInteresting bool // Save games with unusual outcomes
}

// Usage
if config.Verbose {
    log.Printf("Player %d: %s -> Player %d", 
        actor.ID, action.Name(), target.ID)
}
```

### Reproducibility
```go
// Always log seed for failed games
func RunGame(seed int64) (result Result, err error) {
    defer func() {
        if r := recover(); r != nil {
            err = fmt.Errorf("game crashed with seed %d: %v", seed, r)
        }
    }()
    
    rand.Seed(seed)
    // Run game...
}
```

### Interesting Game Detection
```go
// Save games that are statistical outliers
func IsInteresting(result GameResult) bool {
    return result.Turns > 100 ||  // Very long game
           result.Turns < 5 ||     // Very short game
           result.WinnerCoins > 20 || // Huge coin accumulation
           result.TotalChallenges > 30 // Challenge-heavy game
}
```

## Statistical Validation

### Ensure Fair Randomness
```go
// Use crypto/rand for initial seeds
import "crypto/rand"

func GetSecureSeed() int64 {
    var seed int64
    binary.Read(rand.Reader, binary.LittleEndian, &seed)
    return seed
}

// Validate shuffle algorithm with chi-squared test
func TestShuffleFairness(t *testing.T) {
    // Each position should have equal probability
    // Run 10,000 shuffles and verify distribution
}
```

### Prevent Bias
```go
// WRONG: Biased target selection
target := players[rand.Intn(len(players))]

// RIGHT: Exclude self, only living players
validTargets := g.GetValidTargets(actor)
if len(validTargets) == 0 {
    return ErrNoValidTargets
}
target := validTargets[rand.Intn(len(validTargets))]
```

## Command-Line Interface
```go
// main.go structure
var (
    games     = flag.Int("games", 1000000, "Total games to simulate")
    workers   = flag.Int("workers", runtime.NumCPU(), "Parallel workers")
    verbose   = flag.Bool("v", false, "Verbose output")
    profile   = flag.String("profile", "", "Enable profiling (cpu|mem)")
    seed      = flag.Int64("seed", 0, "Random seed (0 for random)")
    output    = flag.String("output", "./results", "Output directory")
)

func main() {
    flag.Parse()
    
    if *profile == "cpu" {
        defer profile.Start().Stop()
    } else if *profile == "mem" {
        defer profile.Start(profile.MemProfile).Stop()
    }
    
    // Run simulation...
}
```

## Code Review Checklist
Before considering implementation complete:

### Game Logic
- [ ] All Coup rules correctly implemented
- [ ] Challenge mechanics working properly
- [ ] Block mechanics working properly
- [ ] Forced coup at 10+ coins
- [ ] Proper elimination handling
- [ ] Deck reshuffling when needed

### AI Behavior
- [ ] 30% bluff rate implemented
- [ ] 50% challenge rate implemented
- [ ] Always blocks when possible
- [ ] Random (fair) target selection

### Performance
- [ ] Achieves 1000+ games/second
- [ ] Memory usage stable over long runs
- [ ] Proper concurrent execution
- [ ] No race conditions (run with -race flag)

### Testing
- [ ] All edge cases covered
- [ ] Integration tests pass
- [ ] Benchmark tests included
- [ ] Reproducible with seed

### Output
- [ ] All CSV files generated correctly
- [ ] Statistics accurately calculated
- [ ] Progress reporting works
- [ ] Character strength rankings clear

## Common Pitfalls to Avoid
1. **Don't modify slices while iterating** - Use indices or copy first
2. **Don't share memory between goroutines** - Use channels or mutexes
3. **Don't ignore error returns** - Always handle or explicitly ignore
4. **Don't hardcode magic numbers** - Use named constants
5. **Don't trust user input** - Validate everything
6. **Don't optimize prematurely** - Profile first
7. **Don't forget edge cases** - Test thoroughly

## Quick Reference: Game Flow
```
1. Deal 2 cards to each player, 2 coins each
2. TURN LOOP:
   a. Active player chooses action
   b. Others can challenge (if applicable)
   c. Resolve challenge if any
   d. Others can block (if applicable)  
   e. Blocker can be challenged
   f. Execute action if not stopped
   g. Check eliminations
   h. Check win condition
   i. Next player
3. Declare winner
4. Record metrics
```

## Questions to Ask During Development
- Is this the simplest solution that works?
- Have I tested the edge cases?
- Will this scale to 1 million games?
- Is the randomness truly random?
- Can I reproduce this bug with a seed?
- Am I measuring what I think I'm measuring?
- Would another developer understand this code?

## Final Notes
- Start simple, iterate toward complexity
- Test early, test often
- Profile before optimizing
- Document surprising behavior
- Keep the specification.md as source of truth
- Ask for clarification if rules are ambiguous

