package cmd

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/openarso/arso/apps/internal/satellite"
)

func SelectResolvedTarget(
	in io.Reader,
	out io.Writer,
	candidates []satellite.ResolvedTarget,
) (satellite.ResolvedTarget, error) {
	if len(candidates) == 0 {
		return satellite.ResolvedTarget{}, fmt.Errorf("no candidates to select")
	}

	fmt.Fprintln(out, "Several satellites match your target:")
	fmt.Fprintln(out)

	for i, candidate := range candidates {
		fmt.Fprintf(
			out,
			"%d) %s — NORAD %d — Object ID %s\n",
			i+1,
			candidate.Name,
			candidate.NoradID,
			candidate.ObjectID,
		)
	}

	fmt.Fprintln(out)

	scanner := bufio.NewScanner(in)

	for {
		fmt.Fprintf(out, "Select satellite [1-%d]: ", len(candidates))

		if !scanner.Scan() {
			return satellite.ResolvedTarget{}, fmt.Errorf("read selection: %w", scanner.Err())
		}

		rawInput := strings.TrimSpace(scanner.Text())

		selectedIndex, err := strconv.Atoi(rawInput)
		if err != nil {
			fmt.Fprintln(out, "Please enter a valid number.")
			continue
		}

		if selectedIndex < 1 || selectedIndex > len(candidates) {
			fmt.Fprintf(out, "Please enter a number between 1 and %d.\n", len(candidates))
			continue
		}

		return candidates[selectedIndex-1], nil
	}
}
