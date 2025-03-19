package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
)

type PushMessage struct {
	To       string `json:"to"`
	Title    string `json:"title"`
	Body     string `json:"body"`
	Data     any    `json:"data,omitempty"`
	Sound    string `json:"sound,omitempty"`
	Priority string `json:"priority,omitempty"`
}

func SendPushNotification(token string, text string) error {
	if token == "" {
		return errors.New("push token is empty")
	}

	message := PushMessage{
		To:       token,
		Title:    "فنجون",
		Body:     text,
		Data:     nil,
		Sound:    "default",
		Priority: "high",
	}

	jsonData, err := json.Marshal([]PushMessage{message})
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", "https://exp.host/--/api/v2/push/send", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("failed to send push notification")
	}

	return nil
}
