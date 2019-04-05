package notifications

import (
	"github.com/hooligram/hooligram-server/db"
	"github.com/hooligram/hooligram-server/utils"
)

var (
	notificationChan = make(chan *NotificationRequest)
)

func HandleNotification() {
	for {
		notificationReq := <-notificationChan

		recipientIDs := notificationReq.RecipientIDs
		notificationClient := &NotificationClient{}
		notificationClient.Init()

		for i, recipientID := range recipientIDs {
			recipientToken, err := db.GetTokenByClientID(recipientID)
			if err != nil {
				utils.LogInfo(notificationsTag, err)
			}

			notificationReq.RecipientIDs[i] = recipientToken
		}
	
		notificationClient.Send(notificationReq)
	}
}
