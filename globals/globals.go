package globals

import (
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/hooligram/hooligram-server/structs"
)

var HttpClient = &http.Client{}
var MessageDeliveryChan = make(chan *structs.MessageDelivery)
var TwilioAPIKey string
var Upgrader = websocket.Upgrader{}

// func getClient(conn *websocket.Conn) (*structs.Client, error) {
// 	client, ok := clients[conn]

// 	if !ok {
// 		return nil, errors.New("i couldn't find you")
// 	}

// 	return client, nil
// }

// func GetSignedInClient(conn *websocket.Conn) (*structs.Client, error) {
// 	client, err := getClient(conn)

// 	if err != nil {
// 		return nil, err
// 	}

// 	if !client.IsSignedIn {
// 		return nil, errors.New("you need to sign in first")
// 	}

// 	return client, nil
// }

// func SignIn(
// 	conn *websocket.Conn,
// 	countryCode, phoneNumber, verificationCode string,
// ) (*structs.Client, error) {
// 	client, ok := db.FindVerifiedClient(countryCode, phoneNumber, verificationCode)

// 	if !ok {
// 		return nil, errors.New("couldn't find such verified client")
// 	}

// 	clients[conn].ID = client.ID
// 	clients[conn].CountryCode = client.CountryCode
// 	clients[conn].PhoneNumber = client.PhoneNumber
// 	clients[conn].VerificationCode = verificationCode
// 	clients[conn].DateCreated = client.DateCreated
// 	clients[conn].IsSignedIn = true

// 	return clients[conn], nil
// }

// func SignOut(conn *websocket.Conn) {
// 	delete(clients, conn)
// }

// func UnverifyClient(client *structs.Client, conn *websocket.Conn) error {
// 	return VerifyClient(client, conn, "")
// }

// func VerifyClient(client *structs.Client, conn *websocket.Conn, verificationCode string) error {
// 	err := db.UpdateClientVerificationCode(client, verificationCode)

// 	if err != nil {
// 		delete(clients, conn)
// 		return err
// 	}

// 	clients[conn].VerificationCode = verificationCode

// 	return nil
// }
