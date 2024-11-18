package simple.authz

import rego.v1

__rego__metadoc__ := {
	"id": "PL159",
	"title": "Users can only access documents with id <= 5 or are in the 'admin' group",
	"description": "This policy allows users to access documents with id <= 5 or are in the 'admin' group",
	"custom": {
		"severity": "High",
		"controls": "BAR-FOO_v1.3.0",
	},
}

# METADATA
# description: Allow only admins, or reading public resources
# entrypoint: true
default allow := false

## useful for debugging
# din = input
# users = data.users

# user may only access documents with id <= 5
allow if {
	input.username == "user"
	input.document.id <= 5
}

# Allow if the user is in the "admin" group according to external data
allow if {
	user_is_in_group("admin", data.users, input.username)
}

default amount_allowed := false

amount_allowed if {
	user_is_in_group("admin", data.users, input.username)
}

amount_allowed if {
	input.amount <= data.users[input.username].maxAmount
}

# Helper function to check if user is in a given group
user_is_in_group(group, users, username) if {
	user_groups := users[username].groups

	group in user_groups
}
