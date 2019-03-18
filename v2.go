package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

func v2(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Println("[V2] Failed to upgrade to WebSocket connection.")
		return
	}

	clients[conn] = &Client{}
	defer delete(clients, conn)
	defer conn.Close()

	for {
		action := Action{}
		err = conn.ReadJSON(&action)

		if err != nil {
			log.Println("[V2] Error reading JSON.")
			break
		}

		if action.Type == "" {
			log.Println("[V2] action type is missing")
			continue
		}

		if action.Payload == nil {
			log.Println("[V2] action payload is missing")
			continue
		}

		switch action.Type {
		case authorizationSignInRequest:
			handleAuthorizationSignInRequest(conn, &action)
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
			log.Println("[V2] action type isn't supported")
		}
	}
}

func handleAuthorizationSignInRequest(conn *websocket.Conn, action *Action) {
	countryCode := action.Payload["country_code"].(string)
	phoneNumber := action.Payload["phone_number"].(string)
	verificationCode := action.Payload["code"].(string)
	client, err := signIn(conn, countryCode, phoneNumber, verificationCode)

	if err != nil {
		log.Println("[V2] Couldn't sign in client.")
		log.Println("[V2]", err.Error())
		writeFailure(conn, authorizationSignInFailure, []string{"couldn't sign in you"})
		return
	}

	writeQueuedActions(client)
	action.Type = authorizationSignInSuccess
	client.writeJSON(action)

	undeliveredMessages, err := findUndeliveredMessages(client.ID)
	if err != nil {
		log.Println("failed to find undelivered message ids")
	}

	for _, undeliveredMessage := range undeliveredMessages {
		action := constructDeliverMessageAction(undeliveredMessage)
		client.writeJSON(action)
	}
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
	_memberIds, memberIdsOk := action.Payload["member_ids"].([]interface{})

	memberIds := make([]int, len(_memberIds))
	for i, memberId := range _memberIds {
		memberIds[i] = int(memberId.(float64))
	}

	if !groupNameOk {
		errors = append(errors, "you need to include `name` in payload")
	}
	if !memberIdsOk {
		errors = append(errors, "you need to include `member_ids` in payload")
	}
	if len(memberIds) < 1 {
		errors = append(
			errors,
			"you need to include at least one member in `member_ids` in payload",
		)
	}
	if !containsID(memberIds, client.ID) {
		errors = append(
			errors,
			"you need to include at the group creator in `member_ids` in payload",
		)
	}

	if len(errors) > 0 {
		log.Println(errors)
		writeFailure(conn, verificationSubmitCodeFailure, errors)
		return
	}

	messageGroup, err := createMessageGroup(groupName, memberIds)

	successAction := constructCreateGroupSuccessAction(
		messageGroup.ID,
		messageGroup.Name,
		messageGroup.MemberIDs,
		messageGroup.DateCreated,
	)

	client.writeJSON(successAction)
}
