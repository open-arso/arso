package satellite

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	// QueryCATNR looks up a target by NORAD catalog number.
	QueryCATNR = "CATNR"
	// QueryNAME looks up a target by object name.
	QueryNAME = "NAME"
)

// BuildCelesTrakQuery converts a CLI target into the CelesTrak query key and
// value needed for that lookup.
func BuildCelesTrakQuery(target string) (string, string, error) {
	cleanTarget := strings.TrimSpace(target)

	if cleanTarget == "" {
		return "", "", fmt.Errorf("target cannot be empty")
	}

	if _, err := strconv.Atoi(cleanTarget); err == nil {
		return QueryCATNR, cleanTarget, nil
	}

	return QueryNAME, cleanTarget, nil
}
