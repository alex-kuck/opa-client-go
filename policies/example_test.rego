package example

import future.keywords.if

test_allow_access_if_21 if {
	allow.valid == true with input as {"age": 21}
}

test_allow_access_if_very_very_old if {
	allow.valid == true with input as {"age": 9000}
}

test_deny_access_if_20 if {
	not allow.valid with input as {"age": 20}
}

test_deny_access_if_not_even_born if {
	not allow.valid with input as {"age": -20}
}

test_allow_chuck_norris_with_20 if {
	allow.valid == true with input as {"age": 20, "name": "Chuck Norris"}
}

test_allow_chuck_norris_even_if_not_born_yet if {
	allow.valid == true with input as {"age": -42, "name": "Chuck Norris"}
}
