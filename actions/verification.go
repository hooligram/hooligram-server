package actions

// CreateVerificationRequestCodeFailure .
func CreateVerificationRequestCodeFailure(errors []string) *Action {
	return createFailureAction(VerificationRequestCodeFailure, errors)
}

// CreateVerificationRequestCodeSuccess .
func CreateVerificationRequestCodeSuccess() *Action {
	return createEmptyAction(VerificationRequestCodeSuccess)
}
