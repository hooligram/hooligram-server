package v2

import (
	"github.com/hooligram/hooligram-server/actions"
	"github.com/hooligram/hooligram-server/clients"
)

//////////////////////////////////////
// HANDLER: CONN_KEEP_ALIVE_REQUEST //
//////////////////////////////////////

func handleConnKeepAliveRequest(client *clients.Client, action *actions.Action) *actions.Action {
	success := actions.ConnKeepAliveSuccess()
	client.WriteJSON(success)
	return success
}
