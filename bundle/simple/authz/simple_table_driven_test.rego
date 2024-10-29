package simple.authz_test

import rego.v1

import data.simple.authz

test_admins if {
	tests := [
		{
			"msg": "admin user",
			"input": {"username": "admin"},
			"expected": true,
		},
		{
			"msg": "non-admin user",
			"input": {"username": "user"},
			"expected": false,
		},
	]

	every test in tests {
		result := authz.allow with input as test.input
		result == test.expected
	}
}
