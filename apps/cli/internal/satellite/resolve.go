package satellite

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	QueryCATNR = "CATNR"
	QueryNAME  = "NAME"
)

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
