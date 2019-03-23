package actions

///////////////////////////////
// VERIFICATION_REQUEST_CODE //
///////////////////////////////

// CreateVerificationRequestCodeFailure .
func CreateVerificationRequestCodeFailure(errors []string) *Action {
	return createFailureAction(VerificationRequestCodeFailure, errors)
}

// CreateVerificationRequestCodeSuccess .
func CreateVerificationRequestCodeSuccess() *Action {
	return createEmptyAction(VerificationRequestCodeSuccess)
}

//////////////////////////////
// VERIFICATION_SUBMIT_CODE //
//////////////////////////////

// CreateVerificationSubmitCodeFailure .
func CreateVerificationSubmitCodeFailure(errors []string) *Action {
	return createFailureAction(VerificationSubmitCodeFailure, errors)
}

// CreateVerificationSubmitCodeSuccess .
func CreateVerificationSubmitCodeSuccess() *Action {
	return createEmptyAction(VerificationSubmitCodeSuccess)
}
