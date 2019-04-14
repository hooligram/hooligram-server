package actions

import "github.com/hooligram/hooligram-server/constants"

///////////////////////////
// AUTHORIZATION_SIGN_IN //
///////////////////////////

// AuthorizationSignInFailure .
func AuthorizationSignInFailure(actionID string, errors []string) *Action {
	return constructFailureAction(actionID, constants.AuthorizationSignInFailure, errors)
}

// AuthorizationSignInSuccess .
func AuthorizationSignInSuccess(actionID string) *Action {
	return constructEmptyAction(actionID, constants.AuthorizationSignInSuccess)
}
