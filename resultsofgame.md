# Competitive Level Analysis — 1,000,000 four-player games

Generated 2026-07-02 with: ./coup-game --test-comp 1000000 --seed 42
Engine: post rules-fixes (BUG-11..17) + card-memory AI (FEAT-6) + measured metrics (BUG-18..22).
Levels are assigned uniformly at random per seat, so each level's fair share is ~33.3%.

```

Running competitiveness level analysis...
Testing 1000000 games with mixed AI levels...

COMPETITIVE LEVEL WIN RATES
===========================
Low Competitive:    338286 wins (33.83%)
Medium Competitive: 327184 wins (32.72%)
High Competitive:   334530 wins (33.45%)
Original AI:        0 wins (0.00%)

CHARACTER PREFERENCE WIN RATES BY LEVEL
=====================================

Low Competitive Winners' Preferences:
  Contessa: 80601 occurrences (23.83%)
  Assassin: 77481 occurrences (22.90%)
  Duke: 63646 occurrences (18.81%)
  Captain: 59440 occurrences (17.57%)
  Ambassador: 57118 occurrences (16.88%)

Medium Competitive Winners' Preferences:
  Assassin: 74971 occurrences (22.91%)
  Contessa: 72705 occurrences (22.22%)
  Duke: 65233 occurrences (19.94%)
  Captain: 59051 occurrences (18.05%)
  Ambassador: 55224 occurrences (16.88%)

High Competitive Winners' Preferences:
  Ambassador: 76221 occurrences (22.78%)
  Assassin: 76151 occurrences (22.76%)
  Contessa: 71574 occurrences (21.40%)
  Duke: 57951 occurrences (17.32%)
  Captain: 52633 occurrences (15.73%)

ACTUAL CARDS HELD BY WINNERS
===========================

Low Competitive Winners' Cards:
  Contessa: 104430 occurrences (15.44%)
  Ambassador: 101399 occurrences (14.99%)
  Captain: 101286 occurrences (14.97%)
  Assassin: 97416 occurrences (14.40%)
  Duke: 96603 occurrences (14.28%)

Medium Competitive Winners' Cards:
  Duke: 120578 occurrences (18.43%)
  Assassin: 106402 occurrences (16.26%)
  Captain: 87541 occurrences (13.38%)
  Contessa: 82051 occurrences (12.54%)
  Ambassador: 74133 occurrences (11.33%)

High Competitive Winners' Cards:
  Ambassador: 130058 occurrences (19.44%)
  Duke: 101162 occurrences (15.12%)
  Assassin: 87300 occurrences (13.05%)
  Contessa: 82709 occurrences (12.36%)
  Captain: 77555 occurrences (11.59%)
```
