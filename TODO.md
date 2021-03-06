# TODO

## Now

- reddit feedback:
  - try to hide the fact that panics are used internally a bit
  - show the intended output of the examples, so the people know that shrinking works indeed
  - correctly report state machines with no possible actions
  - docs!
  - HTTP API testing example
- doc.go paragraph about defining properties
- better docs
  - document every exported symbol
  - more examples
  - a tutorial?

## Big things

- failure persistence

## Misc

- make sure our repeat allows shrinker to remove *any* element (no matter if it was a p=1 or p=0.75 one)
- do not play with filter games for the state machine, just find all valid actions
- our functions panic too early which leads to confusing? `go test` output
- extract filter functionality so that it is easy to mix in generators without the baggage like String()
- when generating numbers in range, try to bias based on the min number,
  just like we bias repeat based on the min number?
  - because min number defines the "magnitude" of the whole thing, kind of?
  - so when we are generating numbers in [1000000; +inf) we do not stick with 1000000 too hard
- when shrinking, if we try to lower the wanted bits of some uint64, we have a high chance to draw very low value
  - because high bits will be masked out
  - this can prevent shrinking, when we first lower block A (which e.g. selects the generator), then
    we draw next block B (which the lowered generator wants fewer bits of). Instead of getting a bit value for B
    and doing proper search, we end up getting a small one, and abandoning the generator shrink
- minimize bitstream mis-alignment during minimization (try to make the shape as constant as possible)
  - better, make minimization not care about mis-alignment
- more powerful assume/filter (look at what hypothesis is doing)

## Generators

- floats
- times, durations, locations
- complex numbers
- big numbers
- ip addresses & masks
- deferred
- subset-of-slice
- permutation-of-slice
- runes with rune/range blacklist
- recursive (base + extend)
- arbitrary(type)

## Wild ideas

- "prefix search" shrinking
  - when shrinking, why do we leave the tail the same?
    - we have "misalignment" problems and all that
  - generate random data instead!
    - generate random tails all the time
- recurrence-based generation, because it is hard to stumble upon interesting stuff purely by random
  - start generating already generated stuff, overriding random for some number of draws
    - zip the sequence with itself
  - random jumps of rng, back/forward
  - recurrence-based generation may actually be better than usual fuzzing!
    - because we operate on 64 bits at once, which in most cases correspond to "full value",
      we have e.g. a much better chance to reproduce a multi-byte sequence (exact or slightly altered) somewhere else
      - this is kind-of related to go-fuzz versifier in some way
      - we also can (and should) reuse whole chunks which can correspond to strings/lists/etc.
  - random markov chain which switches states like
    - generate new data
    - reuse existing data starting from
    - reuse existing data altering it like X
  - should transition probabilities be universal or depend on generators?
    - should they also determine where to jump to, so that we jump to "compatible" stuff only?
      - can tag words with compatibility classes
      - can just jump to previous starts of the generated blocks?
  - can explore/exploit trade-off help us decide when to generate random data, and when to reuse existing?
    - probably can do thompson sampling when we have online coverage information
- arbiter-based distributed system tester
