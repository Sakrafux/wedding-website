// Package exports is the use case behind the two admin CSV downloads.
package exports

import (
	"context"

	"github.com/Sakrafux/wedding-website/internal/domain"
	"github.com/Sakrafux/wedding-website/internal/infrastructure/persistence"
)

// UseCase is the use case behind the admin CSV downloads: the code list the print
// shop needs, and the full guest list.
//
// Thin on purpose — the files are projections, and there is no rule to apply to them
// beyond which rows they contain. It exists so the handler talks to a use case rather
// than to a store, which is what keeps SQL out of web.
type UseCase struct {
	households *persistence.HouseholdStore
	export     *persistence.ExportStore
}

func New(households *persistence.HouseholdStore, export *persistence.ExportStore) *UseCase {
	return &UseCase{households: households, export: export}
}

// Codes returns every household with its login code, ordered by name, for
// codes.csv.
//
// The same List the admin screen uses: the printer's file and the screen must agree
// about who exists, and two queries are two chances for them not to.
func (useCase *UseCase) Codes(ctx context.Context) ([]domain.HouseholdOverview, error) {
	return useCase.households.List(ctx)
}

// StreamGuests calls yield once per guest — soft-deleted ones included — with the
// values in persistence.GuestExportColumns order.
func (useCase *UseCase) StreamGuests(ctx context.Context, yield func(values []any) error) error {
	return useCase.export.StreamGuests(ctx, yield)
}

// GuestExportColumns is the header row of guests.csv. Re-exported so the handler does
// not import persistence to write a header.
func (useCase *UseCase) GuestExportColumns() []string {
	return persistence.GuestExportColumns
}
