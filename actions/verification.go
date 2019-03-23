package actions

// CreateVerificationRequestCodeFailure .
func CreateVerificationRequestCodeFailure(errors []string) *Action {
	return createFailureAction(VerificationRequestCodeFailure, errors)
}

// CreateVerificationRequestCodeSuccess .
func CreateVerificationRequestCodeSuccess() *Action {
	return createEmptyAction(VerificationRequestCodeSuccess)
}

// CreateVerificationSubmitCodeFailure .
func CreateVerificationSubmitCodeFailure(errors []string) *Action {
	return createFailureAction(VerificationSubmitCodeFailure, errors)
}

// CreateVerificationSubmitCodeSuccess .
func CreateVerificationSubmitCodeSuccess() *Action {
	return createEmptyAction(VerificationSubmitCodeSuccess)
}
