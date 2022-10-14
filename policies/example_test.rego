package example

import future.keywords.if

test_allow_access_if_21 {
    allow.valid == true with input as {"age": 21}
}

test_allow_access_if_very_very_old {
    allow.valid == true with input as {"age": 9000}
}

test_deny_access_if_20 {
    not allow.valid with input as {"age": 20}
}

test_deny_access_if_not_even_born {
    not allow.valid with input as {"age": -20}
}

test_allow_chuck_norris_with_20 {
    allow.valid == true with input as {"age": 20, "name": "Chuck Norris"}
}

test_allow_chuck_norris_even_if_not_born_yet {
    allow.valid == true with input as {"age": -42, "name": "Chuck Norris"}
}
