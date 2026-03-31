# REVIEW.md — Bug Analysis

## Issue 1: Public struct fields break encapsulation

- **Lines:** 1–8 (struct definition)
- **Problem:** All fields (`ID`, `Name`, `Price`, `Stock`, `Status`) are exported. Any caller can
  bypass business rules and mutate them directly (e.g., `p.Stock = -99`), making invariants
  impossible to enforce.
- **Fix:** Make all fields private (`id`, `name`, `price`, `stock`, `status`) and expose them only
  through accessor methods. Mutations must go through validated methods (`UpdatePrice`,
  `DecreaseStock`, etc.).

---

## Issue 2: `float64` for monetary amounts

- **Line:** 4 (`Price float64`)
- **Problem:** Floating-point arithmetic is inexact for decimal values. A classic example:
  ```
  price := 0.1 + 0.2   // result: 0.30000000000000004, not 0.3
  ```
  Applying a 33.33% discount to $9.99:
  ```
  9.99 * (1 - 33.33/100) = 6.6633... → rounds incorrectly
  ```
  This causes cumulative rounding errors in totals, taxes, and refunds.
- **Fix:** Store price as `int64` in the smallest unit (cents). $9.99 → `999`. All arithmetic
  stays exact.

---

## Issue 3: `context.Context` in a domain method

- **Line:** 10 (`func (p *Product) DecreaseStock(ctx context.Context, ...)`)
- **Problem:** Domain methods model pure business logic. Injecting `context.Context` couples the
  domain to infrastructure concerns (HTTP timeouts, tracing, cancellation). This violates Clean
  Architecture — the domain must have zero infrastructure dependencies.
- **Fix:** Remove `ctx` from the domain method signature entirely. If context propagation is
  needed, handle it in the application-service or repository layer.

---

## Issue 4: Infrastructure (logging) inside the domain method

- **Line:** 21 (`log.Printf("Stock decreased for product %s", p.ID)`)
- **Problem:** Logging is an infrastructure concern. Domain methods must remain pure — no I/O,
  no side-effects other than state mutation. Embedding `log.Printf` here:
  - Creates a hidden dependency on the global logger.
  - Makes unit-testing harder (log output pollutes test runs).
  - Prevents the caller from controlling log format or level.
- **Fix:** Remove the log call from `DecreaseStock`. Log the event at the boundary layer (HTTP
  handler or application service) after the domain method returns successfully.

---

## Issue 5: Weak error message ("not enough")

- **Line:** 14 (`return errors.New("not enough")`)
- **Problem:** The string `"not enough"` is not a sentinel — callers cannot distinguish this error
  from other errors without fragile string comparison. It also carries no context (which product,
  how much was requested vs available).
- **Fix:** Define a package-level sentinel: `var ErrInsufficientStock = errors.New("insufficient stock")`.
  Wrap it in a typed `DomainError` that carries machine-readable context. Callers use
  `errors.Is(err, ErrInsufficientStock)` for branching.

---

## Issue 6: `Status` as plain `string`

- **Line:** 6 (`Status string`)
- **Problem:** Any string can be assigned (`p.Status = "BANANA"`). There is no compile-time or
  runtime validation of allowed values.
- **Fix:** Define a typed constant:
  ```go
  type Status string
  const (
      StatusActive     Status = "ACTIVE"
      StatusInactive   Status = "INACTIVE"
      StatusOutOfStock Status = "OUT_OF_STOCK"
  )
  ```

---

## Issue 7: `ApplyDiscount` has no validation

- **Lines:** 24–26
- **Problem:** Negative or >100% discounts are silently accepted, producing a negative or inflated
  price. There is no error return.
- **Fix:** Validate the percent range (0 ≤ percent ≤ 100) and return an error. Also switch to
  integer arithmetic on cents to avoid the float64 issue from Issue 2.

---

## Issue 8: `ID` is `string`, not a typed ID

- **Line:** 1 (`ID string`)
- **Problem:** A plain `string` ID can be confused with other string fields at call sites,
  e.g., passing a product name where an ID is expected will compile fine.
- **Fix:** Use a typed alias: `type ProductID string`. The compiler will then catch mismatches.
