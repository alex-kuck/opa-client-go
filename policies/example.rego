package example

import future.keywords.if

allow = {"valid": true} if {
	has_legal_age
} else = {"valid": false} if true

has_legal_age if {
	input.age >= 21
} else if {
	input.name == "Chuck Norris"
} else = false if {
	true
}
