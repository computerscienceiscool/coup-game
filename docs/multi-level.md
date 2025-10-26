# Coup Character User Types

This document outlines three competitive profiles (High, Medium, and Low) for each character role in Coup, providing different playstyles and strategic approaches.

## 1. Duke

### High Competitive Duke
- **Playstyle**: Aggressively takes Tax (3 coins) almost every turn, regardless of having the Duke
- **Bluffing**: Strategically blocks Foreign Aid even when not holding Duke (~70% of opportunities)
- **Challenge Behavior**: Challenges others claiming Duke (direct competition)
- **Strategy**: Accumulates coins rapidly for early Coups while controlling opponents' economy
- **AI Parameters**: BluffRate=0.7, ChallengeRate=0.7 for Duke claims

### Medium Competitive Duke
- **Playstyle**: Uses Tax when holding Duke, occasionally bluffs
- **Bluffing**: Blocks Foreign Aid when holding Duke, sometimes bluffs (~40%)
- **Challenge Behavior**: Challenges obvious Duke bluffs
- **Strategy**: Balance between coin collection and defensive play
- **AI Parameters**: BluffRate=0.4, ChallengeRate=0.5

### Low Competitive Duke
- **Playstyle**: Only uses Tax when actually holding Duke
- **Bluffing**: Rarely blocks Foreign Aid unless holding Duke
- **Challenge Behavior**: Seldom challenges Duke claims
- **Strategy**: Focuses on honest play, misses opportunity to control economy
- **AI Parameters**: BluffRate=0.1, ChallengeRate=0.3

## 2. Assassin

### High Competitive Assassin
- **Playstyle**: Aggressively collects exactly 3 coins and assassinates frequently
- **Targeting**: Prioritizes the strongest player or those with Contessa claims
- **Bluffing**: Claims Assassin even when not holding (~60%)
- **Strategy**: Creates fear through assassination threats, forces Contessa reveals
- **AI Parameters**: BluffRate=0.6, ChallengeRate=0.6 for Contessa blocks

### Medium Competitive Assassin
- **Playstyle**: Uses Assassination when appropriate, not every opportunity
- **Targeting**: Chooses targets based on threat level
- **Bluffing**: Sometimes claims Assassin as a bluff (~35%)
- **Strategy**: Uses assassination as one tool among many
- **AI Parameters**: BluffRate=0.35, ChallengeRate=0.5

### Low Competitive Assassin
- **Playstyle**: Only assassinates when clearly advantageous
- **Targeting**: Random or obvious targets
- **Bluffing**: Rarely bluffs with Assassin (~15%)
- **Strategy**: Prefers safer actions, uses assassination conservatively
- **AI Parameters**: BluffRate=0.15, ChallengeRate=0.3

## 3. Captain

### High Competitive Captain
- **Playstyle**: Frequently steals from coin-rich players
- **Bluffing**: Claims Captain to steal even when not holding (~65%)
- **Blocking**: Always blocks steal attempts against self, regardless of having Captain/Ambassador
- **Strategy**: Focuses on disrupting others' economy while building own
- **AI Parameters**: BluffRate=0.65, ChallengeRate=0.6 for steal blocks

### Medium Competitive Captain
- **Playstyle**: Steals when advantageous, not every turn
- **Bluffing**: Sometimes claims Captain as bluff (~40%)
- **Blocking**: Blocks steals when holding Captain/Ambassador or when critical
- **Strategy**: Uses stealing opportunistically
- **AI Parameters**: BluffRate=0.4, ChallengeRate=0.5

### Low Competitive Captain
- **Playstyle**: Only steals when holding Captain
- **Bluffing**: Rarely bluffs with Captain (~10%)
- **Blocking**: Only blocks when actually holding Captain/Ambassador
- **Strategy**: Uses stealing inconsistently, misses opportunities
- **AI Parameters**: BluffRate=0.1, ChallengeRate=0.3

## 4. Ambassador

### High Competitive Ambassador
- **Playstyle**: Uses Exchange strategically to find specific characters
- **Bluffing**: Claims Ambassador for card advantage (~50%)
- **Blocking**: Always blocks steals, regardless of having Ambassador/Captain
- **Strategy**: Uses information advantage from seeing cards, defensive focus
- **AI Parameters**: BluffRate=0.5, ChallengeRate=0.6 for steal actions

### Medium Competitive Ambassador
- **Playstyle**: Uses Exchange when holding Ambassador or needing new cards
- **Bluffing**: Sometimes claims Ambassador (~30%)
- **Blocking**: Blocks steals when holding Captain/Ambassador or when critical
- **Strategy**: Uses card cycling as one tool among many
- **AI Parameters**: BluffRate=0.3, ChallengeRate=0.5

### Low Competitive Ambassador
- **Playstyle**: Rarely uses Exchange even when holding Ambassador
- **Bluffing**: Almost never bluffs with Ambassador (~5%)
- **Blocking**: Only blocks when actually holding Ambassador/Captain
- **Strategy**: Doesn't optimize card selection, plays cards as drawn
- **AI Parameters**: BluffRate=0.05, ChallengeRate=0.3

## 5. Contessa

### High Competitive Contessa
- **Playstyle**: Defensive master, focuses on survival
- **Bluffing**: Always claims Contessa when assassinated (~80% when not holding)
- **Challenge Behavior**: Carefully tracks Assassin claims to challenge accurately
- **Strategy**: Outlasts opponents through defensive play, wins late game
- **AI Parameters**: BluffRate=0.8, ChallengeRate=0.6 for Assassin claims

### Medium Competitive Contessa
- **Playstyle**: Balances defense with offensive actions
- **Bluffing**: Often claims Contessa when assassinated (~50% when not holding)
- **Challenge Behavior**: Challenges obvious Assassin bluffs
- **Strategy**: Uses Contessa defensively while building coins for actions
- **AI Parameters**: BluffRate=0.5, ChallengeRate=0.5

### Low Competitive Contessa
- **Playstyle**: Inconsistent defense, doesn't optimize Contessa use
- **Bluffing**: Only claims Contessa when holding (~15% bluff rate)
- **Challenge Behavior**: Rarely challenges Assassin claims
- **Strategy**: Often loses influence to assassinations, doesn't bluff effectively
- **AI Parameters**: BluffRate=0.15, ChallengeRate=0.3
