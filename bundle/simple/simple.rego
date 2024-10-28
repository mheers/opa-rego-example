package simple.authz

import future.keywords.if
import future.keywords.in

default allow = false
default amountAllowed = false

## useful for debugging
# din = input
# users = data.users

# user may only access documents with id <= 5
allow {
    input.username == "user"
    input.document.id <= 5
}

# Allow if the user is in the "admin" group according to external data
allow {
    user_is_in_group("admin")
}

amountAllowed {
    user_is_in_group("admin")
}

amountAllowed {
    input.amount <= data.users[input.username].maxAmount
}

# Helper function to check if user is in a given group
user_is_in_group(group) if {
    userGroups := data.users[input.username].groups

    # printing user groups
    print("User groups:", userGroups)

    group in userGroups
}
