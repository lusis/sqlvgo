package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFilterGoResults(t *testing.T) {
	results := []*Record{
		// these should match
		&Record{Rstate: 1, Rtype: 1},
		&Record{Rstate: 3, Rtype: 3},
		&Record{Rstate: 1, Rtype: 3},
		&Record{Rstate: 3, Rtype: 1},
		// these should not match
		&Record{Rstate: 2, Rtype: 3},
		&Record{Rstate: 2, Rtype: 1},
		&Record{Rstate: 1, Rtype: 2},
		&Record{Rstate: 3, Rtype: 2},
	}
	gores := filterGoResults(results, []int{1, 3})
	require.Len(t, gores, 4)
}
