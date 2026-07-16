package satellite

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

func normalizeTarget(target string) string {
	return strings.ToUpper(strings.TrimSpace(target))
}

func parseNORADID(target string) (int, bool) {
	noradID, err := strconv.Atoi(strings.TrimSpace(target))
	if err != nil {
		return 0, false
	}

	if noradID <= 0 {
		return 0, false
	}

	return noradID, true
}

// ResolveTarget resolves a user-supplied target string to a single satellite,
// preferring direct NORAD IDs and cached lookups before calling CelesTrak.
func (c *Client) ResolveTarget(ctx context.Context, target string) (ResolvedTarget, error) {
	normalized := normalizeTarget(target)

	if normalized == "" {
		return ResolvedTarget{}, fmt.Errorf("target cannot be empty")
	}

	if noradID, ok := parseNORADID(normalized); ok {
		return ResolvedTarget{
			Query:   target,
			Name:    normalized,
			NoradID: noradID,
			Kind:    "satellite",
			Source:  "numeric",
		}, nil
	}

	if c.targetCache != nil {
		if resolved, ok := c.targetCache.GetTarget(normalized); ok {
			resolved.Source = "cache"
			return resolved, nil
		}
	}

	resolved, err := c.resolveTargetFromCelesTrak(ctx, normalized)
	if err != nil {
		return ResolvedTarget{}, err
	}

	if c.targetCache != nil {
		_ = c.targetCache.SetTarget(normalized, resolved)
	}

	return resolved, nil
}

func (c *Client) resolveTargetFromCelesTrak(ctx context.Context, target string) (ResolvedTarget, error) {
	body, err := c.fetchRaw(ctx, QueryNAME, target)
	if err != nil {
		return ResolvedTarget{}, err
	}

	var elements []GPElement
	if err := json.Unmarshal(body, &elements); err != nil {
		return ResolvedTarget{}, fmt.Errorf("decode CelesTrak response: %w", err)
	}

	if len(elements) == 0 {
		return ResolvedTarget{}, fmt.Errorf("no satellite found for %q", target)
	}

	if len(elements) > 1 {
		candidates := make([]ResolvedTarget, 0, len(elements))

		for _, element := range elements {
			candidates = append(candidates, ResolvedTarget{
				Query:    target,
				Name:     element.ObjectName,
				ObjectID: element.ObjectID,
				NoradID:  element.NoradCatID,
				Kind:     "satellite",
				Source:   "celestrak",
			})
		}

		return ResolvedTarget{}, &AmbiguousTargetError{
			Target:     target,
			Candidates: candidates,
		}
	}

	element := elements[0]

	now := time.Now().UTC()

	return ResolvedTarget{
		Query:      target,
		Name:       element.ObjectName,
		ObjectID:   element.ObjectID,
		NoradID:    element.NoradCatID,
		Kind:       "satellite",
		Source:     "celestrak",
		ResolvedAt: now,
		ExpiresAt:  now.Add(30 * 24 * time.Hour),
	}, nil
}
