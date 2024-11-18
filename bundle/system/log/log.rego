package system.log

import rego.v1

# If the name field of input is marcel, remove the salary field
mask contains "/input/salary" if {
	input.input.username == "marcel"
}

# Remove input's password field
mask contains "/input/password"

# If a card field exists in input, change the value to ****-****-****-****
mask contains {"op": "upsert", "path": "/input/card", "value": x} if {
	input.input.card
	x := "****-****-****-****"
}
