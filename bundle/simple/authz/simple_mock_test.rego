package simple.authz_test

import rego.v1

import data.simple.authz

mock_users := {
	"admin": {"groups": ["admin"]},
	"john": {"groups": ["admin"]},
	"user": {"groups": ["user"]},
}

test_admins_mock if {
	tests := [
		{
			"msg": "admin user",
			"username": "admin",
			"expected": true,
		},
		{
			"msg": "non-admin user",
			"username": "user",
			"expected": false,
		},
		{
			"msg": "john is admin",
			"username": "john",
			"expected": true,
		},
	]

	every test in tests {
		result := authz.allow with input as {"username": test.username}
			with data.users as mock_users
		result == test.expected
	}
}
