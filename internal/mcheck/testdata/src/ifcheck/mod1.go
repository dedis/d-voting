package ifcheck

func dummy() {
	// Should be good
	a := true
	if a {

	}

	// Should raise a concern
	if b := false; b { // want "Please do not do initialization in if statement"

	}

	// Should work if nested
	for i := 0; i < 1; i++ {
		// No concern
		if i == 2 {

		}
		// Should raise a concern
		if c := true; c { // want "Please do not do initialization in if statement"

		}
	}
}
