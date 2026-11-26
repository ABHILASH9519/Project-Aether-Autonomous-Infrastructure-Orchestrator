package lifecycle

import (
	"errors"
	"testing"
)

func TestLifecycle(t *testing.T) {
	state := "open"
	for _, tc := range []struct{ action, want string; steps int }{
		{"acknowledge", "acknowledged", 1},
		{"request_approval", "awaiting_approval", 1},
		{"approve", "approved", 1},
		{"simulate", "resolved", 3},
	} {
		steps, err := Plan(state, tc.action)
		if err != nil || len(steps) != tc.steps || steps[len(steps)-1].State != tc.want {
			t.Fatalf("%s/%s: %v %v", state, tc.action, steps, err)
		}
		state = tc.want
	}
}

func TestApprovalCannotBeBypassed(t *testing.T) {
	for _, state := range []string{"open", "acknowledged", "awaiting_approval", "rejected", "resolved"} {
		if _, err := Plan(state, "simulate"); !errors.Is(err, ErrTransition) {
			t.Fatalf("simulation unexpectedly allowed from %s", state)
		}
	}
}
func TestRejection(t *testing.T) {
	steps, err := Plan("awaiting_approval", "reject")
	if err != nil || len(steps) != 1 || steps[0].State != "rejected" { t.Fatal(steps, err) }
}
