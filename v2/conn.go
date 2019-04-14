package v2

import (
	"github.com/hooligram/hooligram-server/actions"
	"github.com/hooligram/hooligram-server/clients"
)

//////////////////////////////////////
// HANDLER: CONN_KEEP_ALIVE_REQUEST //
//////////////////////////////////////

func handleConnKeepAliveRequest(client *clients.Client, action *actions.Action) *actions.Action {
	actionID := action.ID
	if actionID == "" {
		failure := actions.ConnKeepAliveFailure(actionID)
		client.WriteJSON(failure)
		return failure
	}

	success := actions.ConnKeepAliveSuccess(actionID)
	client.WriteJSON(success)
	return success
}
