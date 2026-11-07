package evaluate

func allStates(left, right State) State {
	if left == False || right == False {
		return False
	}
	if left == True && right == True {
		return True
	}
	return Unknown
}

// anyStates and negate implement Firebase's supported audience membership
// methods, independently of the AND-only condition expression grammar.
func anyStates(left, right State) State {
	if left == True || right == True {
		return True
	}
	if left == False && right == False {
		return False
	}
	return Unknown
}

func negate(state State) State {
	switch state {
	case True:
		return False
	case False:
		return True
	default:
		return Unknown
	}
}
