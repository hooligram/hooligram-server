package v2

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/websocket"
	"github.com/hooligram/hooligram-server/actions"
	"github.com/hooligram/hooligram-server/clients"
	"github.com/hooligram/hooligram-server/constants"
	"github.com/hooligram/hooligram-server/utils"
)

const v2Tag = "v2"

var upgrader = websocket.Upgrader{}

// V2 .
func V2(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		utils.LogInfo(v2Tag, "error upgrading to websocket. "+err.Error())
		return
	}

	clients.Add(conn)
	defer clients.Remove(conn)
	defer conn.Close()

	for {
		client, ok := clients.Get(conn)
		if !ok {
			break
		}

		var p []byte
		_, p, err = conn.ReadMessage()
		if err != nil {
			utils.LogInfo(
				v2Tag,
				fmt.Sprintf("connection error. client id %v. %v", client.GetID(), err.Error()),
			)
			return
		}

		action := actions.Action{}
		err = json.Unmarshal(p, &action)

		if err != nil {
			utils.LogInfo(
				v2Tag,
				fmt.Sprintf("error reading json. client id %v. %v", client.GetID(), err.Error()),
			)
			continue
		}

		if action.Type == "" {
			utils.LogInfo(
				v2Tag,
				fmt.Sprintf("action type missing. client id %v. %v", client.GetID(), err.Error()),
			)
			continue
		}

		if action.Payload == nil {
			utils.LogInfo(
				v2Tag,
				fmt.Sprintf("action payload missing. client id %v. %v", client.GetID(), err.Error()),
			)
			continue
		}

		utils.LogOpen(client.SessionID, strconv.Itoa(client.GetID()), action.Type, action.Payload)
		var result *actions.Action

		switch action.Type {
		case constants.AuthorizationSignInRequest:
			result = handleAuthorizationSignInRequest(client, &action)
		case constants.GroupAddMemberRequest:
			result = handleGroupAddMemberRequest(client, &action)
		case constants.GroupCreateRequest:
			result = handleGroupCreateRequest(client, &action)
		case constants.GroupDeliverSuccess:
			result = handleGroupDeliverSuccess(client, &action)
		case constants.GroupLeaveRequest:
			result = handleGroupLeaveRequest(client, &action)
		case constants.MessagingSendRequest:
			result = handleMessagingSendRequest(client, &action)
		case constants.MessagingDeliverSuccess:
			result = handleMessagingDeliverSuccess(client, &action)
		case constants.VerificationRequestCodeRequest:
			result = handleVerificationRequestCodeRequest(client, &action)
		case constants.VerificationSubmitCodeRequest:
			result = handleVerificationSubmitCodeRequest(client, &action)
		default:
		}

		if result == nil {
			continue
		}

		utils.LogClose(client.SessionID, strconv.Itoa(client.GetID()), result.Type, result.Payload)
	}
}

func generateRandomActionID() string {
	return utils.GenerateRandomString(constants.ActionIDLength)
}
