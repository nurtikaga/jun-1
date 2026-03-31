# ANSWERS.md

## Q1 — What happens when `DecreaseStock(0)` is called?

Calling `DecreaseStock(0)` is treated as a **no-op and returns `nil`**.

This is the correct behaviour. A zero-quantity decrease means "take nothing away from stock,"
which is a valid, idempotent operation. Returning an error for quantity=0 would force every
caller to guard the call with `if quantity > 0`, adding noise without safety benefit. The change
tracker is intentionally not updated, because nothing actually changed.

What would be wrong is treating it as an error: clients retrying a request (network blip,
at-least-once delivery) should be able to call `DecreaseStock(0)` safely without triggering
spurious failures.

---

## Q2 — Why is `float64` problematic for prices?

IEEE-754 double-precision floating point cannot represent most decimal fractions exactly.
`0.1` and `0.2` have no exact binary representation, so:

```
0.1 + 0.2 = 0.30000000000000004   (not 0.30)
```

A real-world failure: a checkout system sums 3 items priced at $0.10 each:

```
0.10 + 0.10 + 0.10 = 0.30000000000000004
```

Rounded to 2 decimal places this works, but with larger amounts, cascading rounding errors
accumulate across thousands of transactions, leading to balance discrepancies that are extremely
difficult to audit.

**Fix:** Store price as `int64` in cents. `$9.99 → 999`. All arithmetic is exact integer
arithmetic; format as a decimal only for display.

---

## Q3 — Why does ChangeTracker use `Track()` explicitly instead of auto-detecting changes?

Auto-detection by comparing old vs new values sounds attractive but has several problems:

1. **Equality semantics are type-specific.** Structs, slices, and custom types need custom
   comparators. A generic `reflect.DeepEqual` approach is slow and error-prone.

2. **A field set to the same value is still intentional.** If a caller explicitly calls
   `UpdatePrice(currentPrice)`, they may want the persistence layer to issue an UPDATE
   (e.g., to bump an `updated_at` timestamp). Auto-detection would silently swallow this.

3. **Explicit tracking is a deliberate design signal.** Every mutation in the domain calls
   `tracker.Track(FieldX, newValue)`. This makes the code self-documenting: any reader knows
   exactly which fields can change and when. Auto-detection hides that contract.

4. **Simpler, faster, easier to test.** The `dirty` map is populated only when a mutation
   succeeds, so the tracker can never be in an inconsistent state. Unit tests check
   `Dirty(FieldX)` directly with no reflection overhead.
