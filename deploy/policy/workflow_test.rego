package aether.workflow_test

import rego.v1
import data.aether.workflow.allow

test_operator_can_acknowledge if {
	allow with input as {"mode":"simulation","role":"operator","action":"acknowledge"}
}
test_operator_cannot_approve if {
	not allow with input as {"mode":"simulation","role":"operator","action":"approve"}
}
test_approver_can_approve if {
	allow with input as {"mode":"simulation","role":"approver","action":"approve"}
}
test_approver_cannot_execute if {
	not allow with input as {"mode":"simulation","role":"approver","action":"simulate"}
}
test_live_mode_denied if {
	not allow with input as {"mode":"live","role":"operator","action":"simulate"}
}
test_unknown_action_denied if {
	not allow with input as {"mode":"simulation","role":"operator","action":"shell"}
}

test_operator_can_enqueue_lab if {
	allow with input as {"mode":"simulation","role":"operator","action":"execute_lab"}
}
test_approver_cannot_enqueue_lab if {
	not allow with input as {"mode":"simulation","role":"approver","action":"execute_lab"}
}
test_live_lab_denied if {
	not allow with input as {"mode":"live","role":"operator","action":"execute_lab"}
}
