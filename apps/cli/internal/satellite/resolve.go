package satellite

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	QueryCATNR = "CATNR"
	QueryNAME  = "NAME"

	ISSNoradCatalogID = "25544"
)

func ResolveTarget(target string) (string, string, error) {
	cleanTarget := strings.TrimSpace(target)

	if cleanTarget == "" {
		return "", "", fmt.Errorf("target cannot be empty")
	}

	if strings.EqualFold(cleanTarget, "ISS") {
		return QueryCATNR, ISSNoradCatalogID, nil
	}

	if _, err := strconv.Atoi(cleanTarget); err == nil {
		return QueryCATNR, cleanTarget, nil
	}

	return QueryNAME, cleanTarget, nil
}
