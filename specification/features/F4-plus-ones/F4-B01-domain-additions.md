# `F4-B01` — Domain: the plus-one rule, deletion rules, `guest_added` origin

**Epic:** F4 — Plus-one · **Layer:** backend · **Depends on:** `F3-B01`

## Story

As a developer, I want "who may add whom" as a pure function, so that the one case we automate — a guest invited alone bringing somebody — cannot quietly widen into a household inviting four people we have never heard of.

## Scope

**In:**

- `CanHouseholdAddPlusOne([]Guest) error` — the whole rule, over the household's living members.
- `ErrPlusOneNotAllowed`, and the reason it carries.
- `CanHouseholdRemove(Guest) error` — a household removes only what it added.
- Removing `Settings.DefaultAdditionLimit` and the `app_setting` row behind it, in migration `0003`.

**Out:**

- Endpoints → `F4-B02`, `F4-B03`. Deadline → `F3-B04`, whose option both reuse.
- The admin's own add and remove → `F5-B02`, which has no rule at all beyond validity. That asymmetry is the design, not an omission.

## Instructions

1. The rule, in full: a household may add **one adult**, and only if its living members are **exactly one person**. No children, no second addition, no addition to a household of two or more. Everything else goes through us, and the form says so with a phone number (`F4-F01`).
2. Tightened on 2026-08-31 from a numeric cap of two. Record why, because the looser version is the obvious thing for a later reader to restore: the single genuine unknown at the moment we address the envelopes is whether a guest invited alone is bringing somebody. Every other addition — a child, a second companion, a friend of a couple — lands on the caterer's count, the seating plan and the budget, and we would rather be told than discover it. The cost is accepted and real: households have almost no control over their own member list.
3. `kind` is not an input on the guest path at all. A plus-one is an adult by construction, so `F4-B02` takes a name and nothing else, and `age` never appears. A child a household wants added is a phone call — which is also how we learn the age the caterer bills against.
4. After the addition the household has two members, so the rule refuses the next one without needing to count additions. The limit is **structural**, and that is why there is no number: `Settings.DefaultAdditionLimit` and the `default_addition_limit` row are removed here, in migration `0003`, along with the parsing in `persistence.Settings` and the fixture in `tests/integration/schema_test.go`. A setting nothing reads is a setting somebody eventually changes and expects to matter.
5. `ErrPlusOneNotAllowed` is one sentinel, not three. The endpoint's German sentence is the same in every case — ring us — so distinguishing "your household is too big" from "you already added somebody" buys a message nobody needs and tells a stranger the shape of a household they cannot see.
6. A soft-deleted member does not count toward the household's size. A household of two that removes its own addition is a household of one again, and may add somebody else. This falls out of counting living members and is worth a test, because it is the one path back.
7. `CanHouseholdRemove` returns `ErrNotGuestAdded` for `origin = 'seeded'`. A household never deletes somebody we put on the list — they answer `attending = 'no'`. The seeded member is our record of who we invited and has to survive the answer.
8. Removal is the soft delete `F5-B02` already implements: the person may hold a seat assignment and appears in the audit trail, and `F7` reports the assignment as stale rather than losing it.
9. A guest-added member who has already answered may still be removed. Removing somebody who said yes is a real correction.
10. A plus-one is created with `origin = 'guest_added'`, `kind = 'adult'`, `attending` NULL, `portion = 'full'`, `seating_need = 'normal'`. Not pre-filled with the household's scope: an added person is a new question, and a defaulted `both` is an answer nobody gave.

## Test plan

- [ ] Unit: a household of one may add; a household of two may not; a household of one that already added may not.
- [ ] Unit: a household whose second member is soft-deleted may add again.
- [ ] Unit: seeded members and guest-added members both count toward household size — the rule is about people in the room, not about who typed them.
- [ ] Unit: `CanHouseholdRemove` rejects a seeded member and allows a guest-added one, answered or not.
- [ ] Unit: a new plus-one is `adult`, `guest_added`, unanswered.
- [ ] Integration: migration `0003` removes the `default_addition_limit` row, and `persistence.Settings` no longer reads it.

## Done when

- [ ] The rule exists as one function, no configuration claims to control it, and no guest path can add a child or a second person.
- [ ] Tests above pass; `go test ./...` is green.
- [ ] Checkbox ticked in `README.md`.
