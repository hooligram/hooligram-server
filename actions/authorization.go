package actions

///////////////////////////
// AUTHORIZATION_SIGN_IN //
///////////////////////////

// CreateAuthorizationSignInFailure .
func CreateAuthorizationSignInFailure(errors []string) *Action {
	return createFailureAction(AuthorizationSignInFailure, errors)
}
