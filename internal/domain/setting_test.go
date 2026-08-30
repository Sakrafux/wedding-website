package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// The deadline is the moment editing closes, so the boundary belongs to the
// closed side: a form still open at exactly the advertised deadline is a form
// that outlives the date on the invitation.
func TestRSVPOpenFollowsTheDeadline(t *testing.T) {
	t.Parallel()

	deadline := time.Date(2027, 5, 17, 21, 59, 59, 0, time.UTC)
	settings := Settings{RSVPDeadline: deadline}

	assert.True(t, settings.RSVPOpen(deadline.Add(-time.Second)))
	assert.False(t, settings.RSVPOpen(deadline))
	assert.False(t, settings.RSVPOpen(deadline.Add(time.Second)))
}
