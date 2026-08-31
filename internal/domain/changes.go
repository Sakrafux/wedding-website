package domain

// Changes is the pair of field maps an audit entry carries: what a value was, and
// what it became.
//
// Only fields that actually changed appear, per F1-B06 — the audit log is a history
// and not a second copy of the database, and a full snapshot of every row would
// duplicate personal data into a table nobody ever deletes from. An unchanged field
// in an audit payload is also actively misleading when the log is being read later
// to work out what somebody did.
//
// Keys are the database column names, which are also the JSON field names of the
// API. One vocabulary means a query typed by hand against audit_log needs no
// lookup table to read.
//
// A code must never appear in here. See NewAdminChangeEntry.
type Changes struct {
	// Before is nil for a creation and After is nil for a deletion, so the audit
	// table's columns end up NULL rather than "{}" — `WHERE before IS NULL` then
	// means what a reader expects when picking through the table by hand.
	Before map[string]any
	After  map[string]any
}

// CreatedChanges is the payload for a creation: an after and no before.
func CreatedChanges(fields map[string]any) Changes {
	return Changes{After: fields}
}

// DeletedChanges is the payload for a deletion: a before and no after.
func DeletedChanges(fields map[string]any) Changes {
	return Changes{Before: fields}
}

// IsEmpty reports that nothing changed, which is what lets a no-op PATCH write no
// audit row at all instead of one saying nothing happened.
func (changes Changes) IsEmpty() bool {
	return len(changes.Before) == 0 && len(changes.After) == 0
}

// compare records name when before and after differ.
//
// The maps are allocated on the first difference, so a patch that changes nothing
// leaves an empty Changes rather than two empty maps.
func (changes *Changes) compare(name string, before, after any) {
	if before == after {
		return
	}
	if changes.Before == nil {
		changes.Before = map[string]any{}
		changes.After = map[string]any{}
	}
	changes.Before[name] = before
	changes.After[name] = after
}

// compareOptionalInt is the *int case: comparing the pointers themselves would
// compare addresses and report every write as a change.
func (changes *Changes) compareOptionalInt(name string, before, after *int) {
	changes.compare(name, unwrapInt(before), unwrapInt(after))
}

// compareOptional is the case of a nullable enum — *Attending, *MealChoice. Like
// compareOptionalInt, it exists because comparing the pointers would compare
// addresses and report every write as a change.
//
// A free function rather than a method, because Go methods cannot be generic. It is
// still unexported, so the vocabulary of audit keys stays inside this package.
func compareOptional[T ~string](changes *Changes, name string, before, after *T) {
	changes.compare(name, unwrapString(before), unwrapString(after))
}

// unwrapString turns a nullable enum into a nil-or-string value, so two absent
// values compare equal and an absent one encodes as JSON null.
func unwrapString[T ~string](value *T) any {
	if value == nil {
		return nil
	}
	return string(*value)
}

// unwrapInt turns a *int into a nil-or-number value, so that two absent values
// compare equal and an absent one encodes as JSON null.
func unwrapInt(value *int) any {
	if value == nil {
		return nil
	}
	return *value
}
