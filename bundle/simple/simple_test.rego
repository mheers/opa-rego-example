package simple.authz

import rego.v1

test_allow_user_with_valid_document_id if {
    allow with input as {
        "username": "user",
        "document": {
            "id": 5
        }
    }
}

test_disallow_user_with_invalid_document_id if {
    not allow with input as {
        "username": "user",
        "document": {
            "id": 6
        }
    }
}

test_allow_admin if {
    allow with input as {
        "username": "admin",
    }
}
