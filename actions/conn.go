package actions

import (
	"github.com/hooligram/hooligram-server/constants"
)

/////////////////////
// CONN_KEEP_ALIVE //
/////////////////////

// ConnKeepAliveSuccess .
func ConnKeepAliveSuccess() *Action {
	return constructEmptyAction(constants.ConnKeepAliveSuccess)
}
