// Package lifecycle defines the deterministic incident transition policy.
package lifecycle

import "errors"

var ErrTransition = errors.New("transition is not permitted")

type Step struct {
	State string
	Action string
}

// Plan returns the audited transitions for an operator command.
// Simulation never performs infrastructure operations.
func Plan(state, action string) ([]Step, error) {
	switch {
	case state == "open" && action == "acknowledge":
		return []Step{{"acknowledged", action}}, nil
	case state == "acknowledged" && action == "request_approval":
		return []Step{{"awaiting_approval", action}}, nil
	case state == "awaiting_approval" && action == "approve":
		return []Step{{"approved", action}}, nil
	case state == "awaiting_approval" && action == "reject":
		return []Step{{"rejected", action}}, nil
	case state == "approved" && action == "simulate":
		return []Step{{"executing", "simulate_start"}, {"verifying", "simulate_verify"}, {"resolved", "simulate_complete"}}, nil
	default:
		return nil, ErrTransition
	}
}
