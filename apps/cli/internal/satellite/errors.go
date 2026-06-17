package satellite

import "fmt"

// AmbiguousTargetError reports that a name lookup matched multiple satellites
// and needs user input to choose one.
type AmbiguousTargetError struct {
	Target     string
	Candidates []ResolvedTarget
}

// Error summarizes the ambiguous lookup for user-facing CLI errors.
func (e *AmbiguousTargetError) Error() string {
	return fmt.Sprintf(
		"target %q is ambiguous: %d satellites found",
		e.Target,
		len(e.Candidates),
	)
}
