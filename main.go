package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

//Meeting Schema
// Id
// Title
// Participants
// Start Time
// End Time
// Creation Timestamp

//Person Schema
// Name
// Email
// RSVP (i.e. Yes/No/MayBe/Not Answered)

func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Home Page")
}

func addNewMeeting(w http.ResponseWriter, r *http.Request) {
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017/?readPreference=primary&appname=MongoDB%20Compass&ssl=false"))
	if err != nil {
		log.Fatal(err)
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)
	meetingsDatabase := client.Database("meetingsAPI")
	meetingCollection := meetingsDatabase.Collection("meeting")
	meetingResult, err := meetingCollection.InsertOne(ctx,
		bson.D{
			{"title", "Test"},
			{"participants", bson.A{
				bson.D{
					{"name", "Pratik"},
					{"email", "pratikbaid3@gmail.com"},
					{"RSVP", "YES"},
				},
				bson.D{
					{"name", "Tejas"},
					{"email", "tejas@gmail.com"},
					{"RSVP", "NO"},
				},
			}},
			{"startTime", "10:30"},
			{"endTime", "12:30"},
			{"creationTime", time.Now()},
		},
	)
	if err != nil {
		log.Fatal(err)
	}
	json.NewEncoder(w).Encode(meetingResult)
}

func addParticipantToMeeting(w http.ResponseWriter, r *http.Request) {
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017/?readPreference=primary&appname=MongoDB%20Compass&ssl=false"))
	if err != nil {
		log.Fatal(err)
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)
	meetingsDatabase := client.Database("meetingsAPI")
	meetingCollection := meetingsDatabase.Collection("meeting")
	id, _ := primitive.ObjectIDFromHex("5f8bc9b4454ae1e32b9d4500")
	result, err := meetingCollection.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.D{
			{"$push", bson.D{{"participants",
				bson.D{
					{"name", "Tejas"},
					{"email", "tejasbaid3@gmail.com"},
					{"RSVP", "YES"},
				},
			}}},
		},
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Updated %v Documents!\n", result.ModifiedCount)
	json.NewEncoder(w).Encode(result)
}

func handleRequest() {
	http.HandleFunc("/meeting", addNewMeeting)
	http.HandleFunc("/addParticipant", addParticipantToMeeting)
	log.Fatal(http.ListenAndServe(":8081", nil))
}

func main() {
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017/?readPreference=primary&appname=MongoDB%20Compass&ssl=false"))
	if err != nil {
		log.Fatal(err)
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)
	handleRequest()
}
