package relayto

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Parameters struct {
	UrlEndpoint    string
	DatabaseName   string
	CollectionName string
	UserName       string
	Password       string
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
		UrlEndpoint:    os.Getenv("URL_ENDPOINT"),
		DatabaseName:   os.Getenv("DATABASE_NAME"),
		CollectionName: os.Getenv("COLLECTION_NAME"),
		UserName:       os.Getenv("USERNAME"),
		Password:       os.Getenv("PASSWORD"),
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

func insertEntry(entry *Entry) (*mongo.InsertOneResult, error) {
	parameters := loadParametersFromEnv()

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

	result, err := insertEntry(&entry)
	if err != nil {
		json.NewEncoder(w).Encode(EntryResult{
			Ok:     false,
			Result: "Failed to create entry",
		})

		return
	}

	json.NewEncoder(w).Encode(EntryResult{
		Ok:     true,
		Result: fmt.Sprintf("%v", result.InsertedID.(primitive.ObjectID).Hex()),
	})
}
