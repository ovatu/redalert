package notifiers

import (
	"errors"
	"net/http"
	"net/url"
	"strings"
	"strconv"
)

func init() {
	registerNotifier("twilio_call", NewTwilioCallNotifier)
}

type TwilioCall struct {
	accountSid   string
	authToken    string
	phoneNumbers []string
	twilioNumber string
	failCountAlertThreshold int
}

var NewTwilioCallNotifier = func(config Config) (Notifier, error) {
	if config.Type != "twilio_call" {
		return nil, errors.New("twilio: invalid config type")
	}

	if config.Config["account_sid"] == "" {
		return nil, errors.New("twilio: invalid account_sid")
	}

	if config.Config["auth_token"] == "" {
		return nil, errors.New("twilio: invalid auth_token")
	}

	if config.Config["twilio_number"] == "" {
		return nil, errors.New("twilio: invalid twilio_number")
	}

	fail_count_threshold := 1
	if i, err := strconv.Atoi(config.Config["fail_count_threshold"]); err == nil {
		fail_count_threshold = i
	}
	
	return Notifier(TwilioCall{
		accountSid:   config.Config["account_sid"],
		authToken:    config.Config["auth_token"],
		phoneNumbers: strings.Split(config.Config["notification_numbers"], ","),
		twilioNumber: config.Config["twilio_number"],
		failCountAlertThreshold: fail_count_threshold,
	}), nil
}

func (a TwilioCall) Name() string {
	return "TwilioCall"
}

func (a TwilioCall) ShouldNotify(failCount int) bool {
	return failCount >= a.failCountAlertThreshold
}

func (a TwilioCall) Notify(msg Message) (err error) {

	smsText := msg.DefaultMessage
	for _, num := range a.phoneNumbers {
		err = PhoneCall(a.accountSid, a.authToken, num, a.twilioNumber, smsText)
		if err != nil {
			return
		}
	}

	return nil

}

func PhoneCall(accountSID string, authToken string, to string, from string, body string) error {

	urlStr := "https://api.twilio.com/2010-04-01/Accounts/" + accountSID + "/Calls.json"

	v := url.Values{}
	v.Set("To", to)
	v.Set("From", from)
	v.Set("Url", "http://twimlets.com/message?Message[0]="+url.QueryEscape(body))
	rb := *strings.NewReader(v.Encode())

	client := &http.Client{}
	req, _ := http.NewRequest("POST", urlStr, &rb)
	req.SetBasicAuth(accountSID, authToken)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)
	if !(resp.StatusCode >= 200 && resp.StatusCode < 300) {
		return errors.New("Invalid Twilio status code")
	}
	return err

}
