package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Person struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	RSVP  string `json:"rsvp"`
}

type Meeting struct {
	Title        string `json:"title"`
	Participants Person `json:"participants"`
	StartTime    string `json:"startTime"`
	EndTime      string `json:"endTime"`
}

func meetings(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		addNewMeeting(w, r)
	case "GET":
		participant, ok := r.URL.Query()["participant"]
		if !ok || len(participant[0]) < 1 {
			log.Println("Url Param 'timeFrame' is present")
			getMeetingByTimeFrame(w, r)
		} else {
			log.Println("Url Param 'participant' is present")
			getMeetingByParticipant(w, r)
		}

	}
}

func getMeetingByTimeFrame(w http.ResponseWriter, r *http.Request) {
	layout := "2006-01-02T15:04:05"
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

	start := r.URL.Query()["start"]
	end := r.URL.Query()["end"]

	fromDate, err := time.Parse(layout, start[0])
	toDate, err := time.Parse(layout, end[0])
	log.Println(fromDate)
	log.Println(toDate)
	if err != nil {
		fmt.Println(err)
	}

	cursor, err := meetingCollection.Find(ctx, bson.M{
		"creationTime": bson.M{
			"$gt": fromDate,
			"$lt": toDate,
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	var episodes []bson.M
	if err = cursor.All(ctx, &episodes); err != nil {
		log.Fatal(err)
	}
	fmt.Println(episodes)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(episodes)
}

func getMeetingByParticipant(w http.ResponseWriter, r *http.Request) {
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

	query := r.URL.Query()
	filters := query["participant"]

	cursor, err := meetingCollection.Find(ctx, (bson.M{"participants": bson.M{"$elemMatch": bson.M{"email": filters[0]}}}))

	if err != nil {
		log.Fatal(err)
	}
	var episodes []bson.M
	if err = cursor.All(ctx, &episodes); err != nil {
		log.Fatal(err)
	}
	fmt.Println(episodes)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(episodes)
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

	id := meetingResult.InsertedID
	cursor, err := meetingCollection.Find(ctx, bson.M{"_id": id})
	if err != nil {
		log.Fatal(err)
	}
	var episodes []bson.M
	if err = cursor.All(ctx, &episodes); err != nil {
		log.Fatal(err)
	}
	fmt.Println(episodes)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(episodes)
}

func getMeetingByID(w http.ResponseWriter, r *http.Request) {
	res1 := strings.Split(r.URL.Path, "/")
	docId := res1[2]
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
	id, _ := primitive.ObjectIDFromHex(docId)
	cursor, err := meetingCollection.Find(ctx, bson.M{"_id": id})
	if err != nil {
		log.Fatal(err)
	}
	var episodes []bson.M
	if err = cursor.All(ctx, &episodes); err != nil {
		log.Fatal(err)
	}
	fmt.Println(episodes)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(episodes)
}

func handleRequest() {
	http.HandleFunc("/meeting/", getMeetingByID)
	http.HandleFunc("/meetings", meetings)
	log.Fatal(http.ListenAndServe(":8081", nil))
}

func main() {
	handleRequest()
}
