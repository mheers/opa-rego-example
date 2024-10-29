package simple.authz_test

import rego.v1

import data.simple.authz

test_allow_user_with_valid_document_id if {
	authz.allow with input as {
		"username": "user",
		"document": {"id": 5},
	}
}

test_disallow_user_with_invalid_document_id if {
	not authz.allow with input as {
		"username": "user",
		"document": {"id": 6},
	}
}

test_allow_admin if {
	authz.allow with input as {"username": "admin"}
}
