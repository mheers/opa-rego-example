package simple.authz_test

import rego.v1

import data.simple.authz

test_admins if {
	tests := [
		{
			"msg": "admin user",
			"input": {"email": "admin@heers.it"},
			"expected": true,
		},
		{
			"msg": "non-admin user",
			"input": {"email": "user@heers.it"},
			"expected": false,
		},
	]

	every test in tests {
		result := authz.allow with input as test.input
		result == test.expected
	}
}
