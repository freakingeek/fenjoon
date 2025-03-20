package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

type PushMessage struct {
	To       []string `json:"to"`
	Title    string   `json:"title"`
	Body     string   `json:"body"`
	Sound    string   `json:"sound,omitempty"`
	Priority string   `json:"priority,omitempty"`
}

func SendPushNotification(tokens []string, text string) error {
	if len(tokens) == 0 {
		return errors.New("push token list is empty")
	}

	message := PushMessage{
		To:       tokens,
		Title:    "فنجون",
		Body:     text,
		Sound:    "default",
		Priority: "high",
	}

	jsonData, err := json.Marshal(message)
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

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to send push notification: %s", string(body))
	}

	fmt.Println("Push sent successfully:", string(body))
	return nil
}
