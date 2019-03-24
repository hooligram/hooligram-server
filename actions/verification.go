package actions

import "github.com/hooligram/hooligram-server/constants"

///////////////////////////////
// VERIFICATION_REQUEST_CODE //
///////////////////////////////

// VerificationRequestCodeFailure .
func VerificationRequestCodeFailure(errors []string) *Action {
	return constructFailureAction(constants.VerificationRequestCodeFailure, errors)
}

// VerificationRequestCodeSuccess .
func VerificationRequestCodeSuccess() *Action {
	return constructEmptyAction(constants.VerificationRequestCodeSuccess)
}

//////////////////////////////
// VERIFICATION_SUBMIT_CODE //
//////////////////////////////

// VerificationSubmitCodeFailure .
func VerificationSubmitCodeFailure(errors []string) *Action {
	return constructFailureAction(constants.VerificationSubmitCodeFailure, errors)
}

// VerificationSubmitCodeSuccess .
func VerificationSubmitCodeSuccess() *Action {
	return constructEmptyAction(constants.VerificationSubmitCodeSuccess)
}
