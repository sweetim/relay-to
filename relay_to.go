package relayto

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Parameters struct {
	UrlEndpoint          string
	DatabaseName         string
	CollectionName       string
	UserName             string
	Password             string
	MessengerId          string
	MessengerAccessToken string
}

func (parameters *Parameters) generateMongoUrl() string {
	return fmt.Sprintf(
		"mongodb+srv://%s:%s@%s/%s?retryWrites=true&w=majority",
		parameters.UserName,
		parameters.Password,
		parameters.UrlEndpoint,
		parameters.DatabaseName)
}

func loadParametersFromEnv() Parameters {
	return Parameters{
		UrlEndpoint:          os.Getenv("URL_ENDPOINT"),
		DatabaseName:         os.Getenv("DATABASE_NAME"),
		CollectionName:       os.Getenv("COLLECTION_NAME"),
		UserName:             os.Getenv("USERNAME"),
		Password:             os.Getenv("PASSWORD"),
		MessengerId:          os.Getenv("MESSENGER_ID"),
		MessengerAccessToken: os.Getenv("MESSENGER_ACCESS_TOKEN"),
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

func toMongoDB(parameters *Parameters, entry *Entry) (*mongo.InsertOneResult, error) {
	client, err := mongo.NewClient(options.Client().ApplyURI(parameters.generateMongoUrl()))
	if err != nil {
		log.Println(err)
		return nil, err
	}

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

	err = client.Connect(ctx)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	defer client.Disconnect(ctx)

	return client.
		Database(parameters.DatabaseName).
		Collection(parameters.CollectionName).
		InsertOne(context.TODO(), entry)
}

const (
	MESSAGING_TYPE_UPDATE = "UPDATE"
	SUCCESSFUL            = "SUCCESSFUL"
)

type Payload struct {
	MessagingType string    `json:"messaging_type"`
	Recipient     Recipient `json:"recipient"`
	Message       Message   `json:"message"`
}
type Recipient struct {
	ID string `json:"id"`
}
type Message struct {
	Text string `json:"text"`
}

func toMessenger(parameters *Parameters, entry *Entry) error {
	data := Payload{
		MessagingType: MESSAGING_TYPE_UPDATE,
		Recipient: Recipient{
			ID: parameters.MessengerId,
		},
		Message: Message{
			Text: entry.Content,
		},
	}

	payloadBytes, err := json.Marshal(data)
	if err != nil {
		log.Println(err)
		return err
	}
	body := bytes.NewReader(payloadBytes)

	req, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("https://graph.facebook.com/v9.0/me/messages?access_token=%s", parameters.MessengerAccessToken),
		body)

	if err != nil {
		log.Println(err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")

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
	log.Println(entry)
	// result, err := toMongoDB(&parameters, &entry)
	err := toMessenger(&parameters, &entry)
	if err != nil {
		json.NewEncoder(w).Encode(EntryResult{
			Ok:     false,
			Result: "Failed to create entry",
		})

		return
	}

	json.NewEncoder(w).Encode(EntryResult{
		Ok:     true,
		Result: SUCCESSFUL,
	})
}
