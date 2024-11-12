# The "system" namespace is reserved for internal use
# by OPA. Authorization policy must be defined under
# system.authz as follows:
package system.authz

import rego.v1

default allow := false # Reject requests by default.

## only allows POSTs to /data/simple/authz or /data/simple/authz/allow
allow if {
	input.method == "POST"
	input.path == ["v1", "data", "simple", "authz"]
}

allow if {
	input.method == "POST"
	input.path == ["v1", "data", "simple", "authz", "allow"]
}

allow if {
	input.method == "POST"
	input.path == ["v1", "data", "simple", "authz", "amount_allowed"]
}
