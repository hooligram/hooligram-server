package actions

import (
	"github.com/hooligram/hooligram-server/constants"
)

/////////////////////
// CONN_KEEP_ALIVE //
/////////////////////

// ConnKeepAliveFailure .
func ConnKeepAliveFailure(actionID string) *Action {
	return constructEmptyAction(actionID, constants.ConnKeepAliveFailure)
}

// ConnKeepAliveSuccess .
func ConnKeepAliveSuccess(actionID string) *Action {
	return constructEmptyAction(actionID, constants.ConnKeepAliveSuccess)
}
