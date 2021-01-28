package relayto

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

type Parameters struct {
	SlackToken   string
	SlackChannel string
}

func loadParametersFromEnv() Parameters {
	return Parameters{
		SlackToken:   os.Getenv("SLACK_TOKEN"),
		SlackChannel: os.Getenv("SLACK_CHANNEL"),
	}
}

type Entry struct {
	Content   string `json:"content"`
	Timestamp uint64 `json:"timestamp"`
}

type EntryResult struct {
	Ok     bool   `json:"ok"`
	Result string `json:"result"`
}

type Payload struct {
	Channel string `json:"channel"`
	Text    string `json:"text"`
}

func toSlack(parameters *Parameters, entry *Entry) error {
	data := Payload{
		Channel: parameters.SlackChannel,
		Text:    entry.Content,
	}

	payloadBytes, err := json.Marshal(data)
	if err != nil {
		log.Println(err)
		return err
	}
	body := bytes.NewReader(payloadBytes)

	req, err := http.NewRequest(
		http.MethodPost,
		"https://slack.com/api/chat.postMessage",
		body)

	if err != nil {
		log.Println(err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", parameters.SlackToken))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println(err)
		return err
	}
	defer resp.Body.Close()

	return nil
}

// RelayToHTTP is a HTTP cloud function
func RelayToHTTP(w http.ResponseWriter, r *http.Request) {
	var entry Entry
	if err := json.NewDecoder(r.Body).Decode(&entry); err != nil {
		log.Println(err)
		json.NewEncoder(w).Encode(EntryResult{
			Ok:     false,
			Result: "Failed to decode body",
		})

		return
	}

	parameters := loadParametersFromEnv()
	err := toSlack(&parameters, &entry)

	if err != nil {
		json.NewEncoder(w).Encode(EntryResult{
			Ok:     false,
			Result: "Failed to relay entry",
		})

		return
	}

	json.NewEncoder(w).Encode(EntryResult{
		Ok:     true,
		Result: "SUCCESSFUL",
	})
}
