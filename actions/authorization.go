package actions

import "github.com/hooligram/hooligram-server/constants"

///////////////////////////
// AUTHORIZATION_SIGN_IN //
///////////////////////////

// AuthorizationSignInFailure .
func AuthorizationSignInFailure(errors []string) *Action {
	return constructFailureAction(constants.AuthorizationSignInFailure, errors)
}
