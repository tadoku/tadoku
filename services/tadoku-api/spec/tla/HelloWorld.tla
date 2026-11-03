---- MODULE HelloWorld ----
\* Proves that the TLC toolchain runs, with one invariant and one liveness
\* property. Delete this model when the first real model lands.
EXTENDS Naturals

VARIABLE count

Init == count = 0

Next == count' = (count + 1) % 4

Spec == Init /\ [][Next]_count /\ WF_count(Next)

CountInRange == count \in 0..3

ReachesThree == <>(count = 3)
====
