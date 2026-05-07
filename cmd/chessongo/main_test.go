package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCLIValidateFEN(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"fen", "4k3/8/8/8/8/8/8/4K3 w - - 0 1"}, &stdout, &stderr)

	require.Equal(t, 0, code)
	require.Contains(t, stdout.String(), "valid")
	require.Empty(t, stderr.String())
}

func TestCLILegalMoves(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"legal", "4k3/8/8/8/8/8/8/4K3 w - - 0 1"}, &stdout, &stderr)

	require.Equal(t, 0, code)
	require.Contains(t, strings.Fields(stdout.String()), "e1e2")
	require.Empty(t, stderr.String())
}

func TestCLIPerft(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"perft", "-depth", "2", "4k3/8/8/8/8/8/8/4K3 w - - 0 1"}, &stdout, &stderr)

	require.Equal(t, 0, code)
	require.Contains(t, stdout.String(), "nodes")
	require.Empty(t, stderr.String())
}
