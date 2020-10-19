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

type Participant struct {
	Name  string
	Email string
	RSVP  string
}

type Meeting struct {
	Title        string
	Participants []Participant
	StartTime    time.Time
	EndTime      time.Time
	CreationTime time.Time
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
		"creationtime": bson.M{
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

	var meeting Meeting
	json.NewDecoder(r.Body).Decode(&meeting)
	meeting.CreationTime = time.Now()
	w.Header().Set("Content-Type", "application/json")

	//Check if the participant is in any other meeting with RSVP yes
	isParticipantClashing := false
	for _, participant := range meeting.Participants {
		if participant.RSVP == "YES" {
			var meetingCheck Meeting
			if err = meetingCollection.FindOne(ctx, bson.M{"starttime": bson.D{{"$lte", meeting.StartTime}}, "endtime": bson.D{{"$gt", meeting.StartTime}}, "participants.email": participant.Email, "participants.rsvp": "YES"}).Decode(&meetingCheck); err != nil {
				if err = meetingCollection.FindOne(ctx, bson.M{"starttime": bson.D{{"$lt", meeting.EndTime}}, "endtime": bson.D{{"$gte", meeting.EndTime}}, "participants.email": participant.Email, "participants.rsvp": "YES"}).Decode(&meetingCheck); err != nil {
					if err = meetingCollection.FindOne(ctx, bson.M{"starttime": bson.D{{"$gte", meeting.StartTime}}, "endtime": bson.D{{"$lte", meeting.EndTime}}, "participants.email": participant.Email, "participants.rsvp": "YES"}).Decode(&meetingCheck); err != nil {
						isParticipantClashing = false
					} else {
						isParticipantClashing = true
					}
				} else {
					isParticipantClashing = true
				}
			} else {
				isParticipantClashing = true
			}
		}
	}
	if isParticipantClashing {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode("Participant is already RSVP'd YES for another meeting in the same time frame")
	} else {
		meetingResult, err := meetingCollection.InsertOne(ctx, meeting)
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
		json.NewEncoder(w).Encode(episodes)
	}

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
