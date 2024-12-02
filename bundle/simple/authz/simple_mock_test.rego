package simple.authz_test

import rego.v1

import data.simple.authz

mock_users := [
	{"email": "admin", "group": ["admin"]},
	{"email": "user", "group": ["user"]},
	{"email": "john", "group": ["admin"]},
]

test_admins_mock if {
	tests := [
		{
			"msg": "admin user",
			"email": "admin",
			"expected": true,
		},
		{
			"msg": "non-admin user",
			"email": "user",
			"expected": false,
		},
		{
			"msg": "john is admin",
			"email": "john",
			"expected": true,
		},
	]

	every test in tests {
		result := authz.allow with input as {"email": test.email}
			with data.users as mock_users
		result == test.expected
	}
}
