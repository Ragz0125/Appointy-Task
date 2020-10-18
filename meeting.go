package main

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client

//Variables for the Database Meeting
type Meeting struct {
	ID           string `json:"id,omitempty" bson:"id,omitempty"`
	Title        string `json:"title,omitempty" bson:"title,omitempty"`
	Participants string `json:"participants,omitempty" bson:"participants,omitempty"`
	StartTime    string `json:"stime,omitempty" bson:"stime,omitempty"`
	EndTime      string `json:"etime,omitempty" bson:"etime,omitempty"`
}

//Variables for the Database Participants
type Participants struct {
	ID    string `json:"id,omitempty" bson:"id,omitempty"`
	Name  string `json:"Name,omitempty" bson:"Name,omitempty"`
	Email string `json:"email,omitempty" bson:"email,omitempty"`
	RSVP  string `json:"rsvp,omitempty" bson:"rsvp,omitempty"`
}

var meetings []Meeting

//Add details to the Database (Using POST Method)
func createMeeting(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	var meeting Meeting
	_ = json.NewDecoder(request.Body).Decode(&meeting)
	collection := client.Database("API").Collection("Meeting")
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	result, _ := collection.InsertOne(ctx, meeting)
	json.NewEncoder(response).Encode(result)
}

//Get the Details of all the available meetings in JSON format
func getMeetings(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	collection := client.Database("API").Collection("Meeting")
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var meeting Meeting
		cursor.Decode(&meeting)
		meetings = append(meetings, meeting)
	}
	if err := cursor.Err(); err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	json.NewEncoder(response).Encode(meetings)
}

//Details of a particular meeting(id) using GET Method
func getMeeting(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	params := mux.Vars(request)
	for _, item := range meetings {
		if item.ID == params["id"] {
			json.NewEncoder(response).Encode(item)
			return
		}
	}
}

//List of meeting within a time-range
func getTiming(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("Content-Type", "application/json")
	var timings []Meeting
	params := mux.Vars(request)
	for _, item := range meetings {
		if (item.StartTime == params["stime"]) && (item.EndTime == params["etime"]) {
			timings = append(timings, item)
		}
	}
	json.NewEncoder(response).Encode(timings)
}

//Details of Meetings of a Participants
func getList(response http.ResponseWriter, request *http.Request) {
	params := mux.Vars(request)
	response.Header().Set("content-type", "application/json")
	collection := client.Database("API").Collection("Participants")
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var p Participants
		cursor.Decode(&p)
		if p.Email == params["email"] {
			json.NewEncoder(response).Encode(p)
		}
	}
}

func main() {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, _ = mongo.Connect(ctx, clientOptions)
	router := mux.NewRouter()
	router.HandleFunc("/meetings", createMeeting).Methods("POST")
	router.HandleFunc("/meetings", getMeetings).Methods("GET")
	router.HandleFunc("/meetings/{id}", getMeeting).Methods("GET")
	router.HandleFunc("/meetings?start=<{stime}>&end=<{etime}>", getTiming).Methods("GET")
	router.HandleFunc("/meetings?participant=<{email}>", getList).Methods("GET")
	http.ListenAndServe(":3000", router)
}
