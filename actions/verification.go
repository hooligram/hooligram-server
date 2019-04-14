package actions

import "github.com/hooligram/hooligram-server/constants"

///////////////////////////////
// VERIFICATION_REQUEST_CODE //
///////////////////////////////

// VerificationRequestCodeFailure .
func VerificationRequestCodeFailure(actionID string, errors []string) *Action {
	return constructFailureAction(actionID, constants.VerificationRequestCodeFailure, errors)
}

// VerificationRequestCodeSuccess .
func VerificationRequestCodeSuccess(actionID string) *Action {
	return constructEmptyAction(actionID, constants.VerificationRequestCodeSuccess)
}

//////////////////////////////
// VERIFICATION_SUBMIT_CODE //
//////////////////////////////

// VerificationSubmitCodeFailure .
func VerificationSubmitCodeFailure(actionID string, errors []string) *Action {
	return constructFailureAction(actionID, constants.VerificationSubmitCodeFailure, errors)
}

// VerificationSubmitCodeSuccess .
func VerificationSubmitCodeSuccess(actionID string) *Action {
	return constructEmptyAction(actionID, constants.VerificationSubmitCodeSuccess)
}
