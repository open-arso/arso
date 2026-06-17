package satellite

import "fmt"

type AmbiguousTargetError struct {
	Target     string
	Candidates []ResolvedTarget
}

func (e *AmbiguousTargetError) Error() string {
	return fmt.Sprintf(
		"target %q is ambiguous: %d satellites found",
		e.Target,
		len(e.Candidates),
	)
}