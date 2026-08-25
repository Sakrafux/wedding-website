// Package dto holds the explicit request and response types of the JSON API.
//
// Domain structs are never serialized directly. The reason is privacy, not purity:
// household.code and household.admin_note live on domain structs and must never
// reach a guest's response. An explicit type per payload makes that class of leak
// impossible rather than merely unlikely.
//
// Where a type deliberately omits a field, the omission is commented — otherwise a
// later reader "fixes" the gap and leaks it.
package dto
