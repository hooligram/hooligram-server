package notifications

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"

)

type Client interface {
	Init()
	Send(*NotificationRequest) (NotificationResponse, error)
}

type NotificationClient struct {
	httpclient *http.Client
	host string
	headers map[string]string
}

type NotificationRequest struct {
	RecipientIDs []string
}

type NotificationResponse struct {
	Body []byte
	Header http.Header
	Status string
}

type Message struct {
	notification Notification `json:"notification"`
	data map[string]interface{} `json:"data"`
	to string `json:"to"`
}

type Notification struct {
	title string `json:"title"`
	body string `json:"body"`
	icon string `json:"icon"`
}

var (
	client NotificationClient
	headers map[string]string
	host string
)

const notificationsTag = "notifications"

func Init(authkey, host string) {
	headers = map[string]string{}
	headers["Authorization"] = authkey
	headers["Content-Type"] = "application/json"
}

func (client *NotificationClient) Init() {
	copyHeaders := map[string]string{}
	for k, v := range headers {
		copyHeaders[k] = v
	}

	client.httpclient = &http.Client{}
	client.host = host
	client.headers = copyHeaders
}

func (client *NotificationClient) Send(req *NotificationRequest) (*NotificationResponse, error) {
	requestbody, err := json.Marshal(&req)
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequest(
		"POST",
		client.host,
		bytes.NewBuffer(requestbody),
	)
	if err != nil {
		return nil, err
	}

	for k, v := range client.headers {
		request.Header.Set(k, v)
	}

	response, err := client.httpclient.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	responsebody, _ := ioutil.ReadAll(response.Body)
	
	return &NotificationResponse{
		Body: responsebody,
		Header: response.Header,
		Status: response.Status,
	}, nil
}
