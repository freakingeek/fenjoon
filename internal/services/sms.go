package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type SendOTPViaSMSRequestBody struct {
	Code      string            `json:"code"`
	Sender    string            `json:"sender"`
	Recipient string            `json:"recipient"`
	Variable  map[string]string `json:"variable"`
}

func SendOTPViaSMS(phone string, otp int) error {

	body := SendOTPViaSMSRequestBody{
		Code:      "bmrlq62kxilkjqk",
		Sender:    "+983000505",
		Recipient: strings.Replace(phone, "0", "+98", 1),
		Variable: map[string]string{
			"otpCode": strconv.Itoa(otp),
		},
	}

	json, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", os.Getenv("SMS_PROVIDER_BASE_URL")+"/v1/sms/pattern/normal/send", bytes.NewBuffer(json))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Apikey", os.Getenv("SMS_PROVIDER_API_KEY"))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to send OTP: received status %d", resp.StatusCode)
	}

	fmt.Println("OTP sent successfully to", phone)
	return nil
}
