package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

const v2Tag = "v2"

func v2(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logInfo(v2Tag, "websocket upgrade failed. "+err.Error())
		return
	}

	clients[conn] = &Client{
		SessionID: generateSessionID(),
		conn:      conn,
	}
	defer delete(clients, conn)
	defer conn.Close()

	for {
		client := clients[conn]

		var p []byte
		_, p, err = conn.ReadMessage()
		if err != nil {
			logInfo(
				v2Tag,
				fmt.Sprintf("connection error. client id %v. %v", client.ID, err.Error()),
			)
			return
		}

		action := Action{}
		err = json.Unmarshal(p, &action)

		if err != nil {
			logInfo(
				v2Tag,
				fmt.Sprintf("error reading json. client id %v. %v", client.ID, err.Error()),
			)
			continue
		}

		if action.Type == "" {
			logInfo(
				v2Tag,
				fmt.Sprintf("action type missing. client id %v. %v", client.ID, err.Error()),
			)
			continue
		}

		if action.Payload == nil {
			logInfo(
				v2Tag,
				fmt.Sprintf("action payload missing. client id %v. %v", client.ID, err.Error()),
			)
			continue
		}

		logOpen(client, &action)
		var result *Action

		switch action.Type {
		case authorizationSignInRequest:
			result = handleAuthorizationSignInRequest(conn, &action)
		case messagingSendRequest:
			handleMessagingSendRequest(conn, &action)
		case messagingDeliverSuccess:
			handleMessagingDeliverSuccess(conn, &action)
		case messagingBroadcastRequest:
			handleMessagingBroadcastRequest(conn, &action)
		case verificationRequestCodeRequest:
			handleVerificationRequestCodeRequest(conn, &action)
		case verificationSubmitCodeRequest:
			handleVerificationSubmitCodeRequest(conn, &action)
		case groupCreateRequest:
			handleGroupCreateRequest(conn, &action)
		default:
		}

		if result != nil {
			logClose(client, &action)
		}
	}
}

func handleAuthorizationSignInRequest(conn *websocket.Conn, action *Action) *Action {
	countryCode := action.Payload["country_code"].(string)
	phoneNumber := action.Payload["phone_number"].(string)
	verificationCode := action.Payload["code"].(string)
	client, err := signIn(conn, countryCode, phoneNumber, verificationCode)

	if err != nil {
		writeFailure(conn, authorizationSignInFailure, []string{"sign in failed"})
		logBody(v2Tag, fmt.Sprintf("couldn't sign in client. %v", err.Error()))
		return &Action{
			Payload: map[string]interface{}{
				"errors": []string{"sign in failed"},
			},
			Type: authorizationSignInFailure,
		}
	}

	writeQueuedActions(client)
	action.Type = authorizationSignInSuccess
	client.writeJSON(action)

	undeliveredMessages, err := findUndeliveredMessages(client.ID)
	if err != nil {
		logBody(v2Tag, "error finding messages to deliver. "+err.Error())
	}

	for _, undeliveredMessage := range undeliveredMessages {
		action := constructDeliverMessageAction(undeliveredMessage)
		client.writeJSON(action)
	}

	return action
}

func handleMessagingSendRequest(conn *websocket.Conn, action *Action) {
	client, err := getSignedInClient(conn)
	if err != nil {
		client.writeFailure(messagingSendFailure, []string{err.Error()})
		return
	}

	content, ok := action.Payload["content"].(string)
	if !ok {
		client.writeFailure(messagingSendFailure, []string{"content is missing"})
		return
	}

	messageGroupID, ok := action.Payload["message_group_id"].(float64)
	if !ok {
		client.writeFailure(messagingSendFailure, []string{"message_group_id is missing"})
		return
	}

	senderID, ok := action.Payload["sender_id"].(float64)
	if !ok {
		client.writeFailure(messagingSendFailure, []string{"sender_id is missing"})
		return
	}

	if client.ID != int(senderID) {
		client.writeFailure(messagingSendFailure, []string{"sender_id mismatch"})
		return
	}

	if !isClientInMessageGroup(int(senderID), int(messageGroupID)) {
		client.writeFailure(messagingSendFailure, []string{"sender doesn't belong to message group"})
		return
	}

	message, err := createMessage(content, int(messageGroupID), int(senderID))
	if err != nil {
		client.writeFailure(messagingSendFailure, []string{err.Error()})
		return
	}

	messageGroupMemberIDs, err := findAllMessageGroupMemberIDs(message.MessageGroupID)
	if err != nil {
		client.writeFailure(messagingSendFailure, []string{err.Error()})
		return
	}

	var recipientIDs []int

	for _, recipientID := range messageGroupMemberIDs {
		if recipientID == int(message.SenderID) {
			continue
		}

		recipientIDs = append(recipientIDs, recipientID)
	}

	for _, recipientID := range recipientIDs {
		createReceipt(message.ID, recipientID)
	}

	messageDeliveryChan <- &MessageDelivery{
		Message:      message,
		RecipientIDs: recipientIDs,
	}

	payload := make(map[string]interface{})
	payload["message_id"] = message.ID
	client.writeJSON(&Action{
		Payload: payload,
		Type:    messagingSendSuccess,
	})
}

func handleMessagingDeliverSuccess(conn *websocket.Conn, action *Action) {
	client, err := getSignedInClient(conn)
	if err != nil {
		client.writeFailure(messagingDeliverFailure, []string{err.Error()})
		return
	}

	messageID, ok := action.Payload["message_id"].(float64)
	if !ok {
		client.writeFailure(messagingDeliverFailure, []string{"message_id is missing"})
		return
	}

	recipientID := client.ID
	err = updateReceiptDateDelivered(int(messageID), recipientID)
	if err != nil {
		client.writeFailure(messagingDeliverFailure, []string{err.Error()})
		return
	}
}

func handleMessagingBroadcastRequest(conn *websocket.Conn, action *Action) {
	client, err := getSignedInClient(conn)

	if err != nil {
		client.writeFailure(messagingBroadcastFailure, []string{err.Error()})
		return
	}

	message, ok := action.Payload["message"].(string)

	if !ok {
		client.writeFailure(messagingBroadcastFailure, []string{"you forgot your message"})
		return
	}

	broadcastChan <- constructBroadcastAction(client, message)
}

func handleVerificationRequestCodeRequest(conn *websocket.Conn, action *Action) {
	errors := []string{}
	countryCode, ok := action.Payload["country_code"].(string)

	if !ok {
		errors = append(errors, "you need to include country_code in payload")
	}

	phoneNumber, ok := action.Payload["phone_number"].(string)

	if !ok {
		errors = append(errors, "you need to include phone_number in payload")
	}

	if len(errors) > 0 {
		writeFailure(conn, verificationRequestCodeFailure, errors)
		return
	}

	client, err := getOrCreateClient(countryCode, phoneNumber)

	if err != nil {
		errors = append(errors, err.Error())
		writeFailure(conn, verificationRequestCodeFailure, errors)
		return
	}

	err = unverifyClient(client, conn)

	if err != nil {
		errors = append(errors, err.Error())
		writeFailure(conn, verificationRequestCodeFailure, errors)
		return
	}

	resp, err := postTwilioVerificationStart(countryCode, phoneNumber)

	if err != nil {
		errors = append(errors, err.Error())
		writeFailure(conn, verificationRequestCodeFailure, errors)
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	if err != nil {
		errors = append(errors, err.Error())
		writeFailure(conn, verificationRequestCodeFailure, errors)
		return
	}

	r := map[string]interface{}{}
	err = json.Unmarshal(body, &r)

	if err != nil {
		errors = append(errors, err.Error())
		writeFailure(conn, verificationRequestCodeFailure, errors)
		return
	}

	if !r["success"].(bool) {
		errors = append(errors, "i failed to make verification start api call")
		writeFailure(conn, verificationRequestCodeFailure, errors)
		return
	}

	writeEmptyAction(conn, verificationRequestCodeSuccess)
}

func handleVerificationSubmitCodeRequest(conn *websocket.Conn, action *Action) {
	var errors []string
	code, ok := action.Payload["code"].(string)

	if !ok {
		errors = append(errors, "you need to include code in payload")
		writeFailure(conn, verificationSubmitCodeFailure, errors)
		return
	}

	client, err := getClient(conn)

	if err != nil {
		errors = append(errors, err.Error())
		writeFailure(conn, verificationSubmitCodeFailure, errors)
		return
	}

	if !client.isVerified() {
		resp, err := getTwilioVerificationCheck(client.CountryCode, client.PhoneNumber, code)

		if err != nil {
			errors = append(errors, err.Error())
			client.writeFailure(verificationSubmitCodeFailure, errors)
			return
		}

		body, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()

		if err != nil {
			errors = append(errors, err.Error())
			client.writeFailure(verificationSubmitCodeFailure, errors)
			return
		}

		r := map[string]interface{}{}
		err = json.Unmarshal(body, &r)

		if err != nil {
			errors = append(errors, err.Error())
			client.writeFailure(verificationSubmitCodeFailure, errors)
			return
		}

		if !r["success"].(bool) {
			unverifyClient(client, conn)
			errors = append(errors, "twilio verification check failed")
			client.writeFailure(verificationSubmitCodeFailure, errors)
			return
		}
	} else {
		if client.VerificationCode != code {
			unverifyClient(client, conn)
			errors = append(errors, "verification code doesn't match")
			client.writeFailure(verificationSubmitCodeFailure, errors)
			return
		}
	}

	verifyClient(client, conn, code)
	writeEmptyAction(conn, verificationSubmitCodeSuccess)
}

func handleGroupCreateRequest(conn *websocket.Conn, action *Action) {
	errors := []string{}

	client, err := getClient(conn)
	if err != nil {
		errors = append(errors, err.Error())
		writeFailure(conn, groupCreateFailure, errors)
		return
	}

	groupName, groupNameOk := action.Payload["name"].(string)
	_memberIDs, memberIDsOk := action.Payload["member_ids"].([]interface{})
	memberIDs := make([]int, len(_memberIDs))

	for i, memberID := range _memberIDs {
		memberIDs[i] = int(memberID.(float64))
	}

	if !groupNameOk {
		errors = append(errors, "you need to include `name` in payload")
	}

	if !memberIDsOk {
		errors = append(errors, "you need to include `member_ids` in payload")
	}

	if len(memberIDs) < 1 {
		errors = append(
			errors,
			"you need to include at least one member in `member_ids` in payload",
		)
	}

	if !containsID(memberIDs, client.ID) {
		errors = append(
			errors,
			"you need to include at the group creator in `member_ids` in payload",
		)
	}

	if len(errors) > 0 {
		log.Println(errors)
		writeFailure(conn, groupCreateFailure, errors)
		return
	}

	messageGroup, err := createMessageGroup(groupName, memberIDs)
	if err != nil {
		writeFailure(conn, groupCreateFailure, errors)
	}

	successAction := constructCreateGroupSuccessAction(
		messageGroup.ID,
		messageGroup.Name,
		messageGroup.MemberIDs,
		messageGroup.DateCreated,
	)

	client.writeJSON(successAction)
}
