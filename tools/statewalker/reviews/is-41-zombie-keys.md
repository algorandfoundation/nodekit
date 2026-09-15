# Review: IS #41 - TUI should strive to prevent zombie key deletion

**Status: confirmed on main (v5.0.x code); reproduction scripted with statewalker**

## The hazard
After a key rotation (`keyreg` to a new participation key), the old key is
marked "not active" by the node immediately, but it keeps voting for the next
~320 rounds (`CatchpointLookback` of effectiveness lag). Deleting it in that
window silently takes the account's stake out of consensus.

## Reproduction (statewalker private network, main branch TUI)
1. `statewalker network up --mode private`, Wallet1 online with key A.
2. Rotate: `goal account addpartkey -a <Wallet1> --roundLastValid 2000000`
   then `goal account changeonlinestatus -a <Wallet1> --online`
   (a rotation helper is a good statewalker follow-up).
3. `goal account listpartkeys` immediately shows the zombie window:
   ```
   Registered  ParticipationID   Last Used  First round  Last round
   yes         IMZR... (new)           N/A         1119     2000000
   no          WWUU... (old)          1134          154     1000154   <-- still voting
   ```
4. TUI keys page renders the paradox directly:
   - `IMZR…` Active=**YES**, Last Vote **N/A** (not effective yet)
   - `WWUU…` Active=**N/A**, Last Vote **1198** (voting *right now*)
5. Selecting `WWUU…` and opening Key Information offers `(d)elete`; pressing `d`
   opens the generic confirm modal ("Are you sure you want to delete this key
   from your node?") with **no indication the key is still voting**. Pressing
   `y` would delete it mid-window.

What the TUI already does right: the *active* key (matching the account's
on-chain participation) only offers `take (o)ffline`, no delete. The gap is
solely the recently-rotated-out key.

## Suggested fix direction
The node cannot tell statelessly when the old key stops being needed, but the
TUI has both keys locally, so a conservative guard is cheap:
- If a key's `Last Vote`/`Last Block Proposal` is within ~320 rounds of the
  current round (or of the newer matching key's `EffectiveFirstValid`), treat
  it as *cooling down*: hide `(d)elete` or replace the generic modal text with
  an explicit "this key is still voting; deleting it will take your stake
  offline until round N" warning (issue's option 2 + 3 combined).
- Simplest safe variant (issue's option 1): only allow deleting keys whose
  `LastValidRound` has passed or that never voted.

`statewalker partkeys zombie` also covers the sibling case (key installed for
an account that never registered it): such keys show Active=N/A and delete
freely. That deletion is harmless, which is why the guard should key off
recent votes / effective rounds, not registration alone.
