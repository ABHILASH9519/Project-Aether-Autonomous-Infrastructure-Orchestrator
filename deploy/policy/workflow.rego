package aether.workflow

import rego.v1

default allow := false

allow if {
	input.mode == "simulation"
	input.role == "operator"
	input.action in {"acknowledge", "request_approval", "simulate"}
}

allow if {
	input.mode == "simulation"
	input.role == "approver"
	input.action in {"approve", "reject"}
}
