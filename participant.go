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

//THIS IS USED TO ADD THE DETAILS OF EACH PARTICIPANT. HERE ID IS THE COMMON KEY BETWEEN THE 2 DATABASES
type Participants struct {
	ID    string `json:"id,omitempty" bson:"id,omitempty"`
	Name  string `json:"Name,omitempty" bson:"Name,omitempty"`
	Email string `json:"email,omitempty" bson:"email,omitempty"`
	RSVP  string `json:"rsvp,omitempty" bson:"rsvp,omitempty"`
}

var ps []Participants

func createParticipant(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	var p Participants
	_ = json.NewDecoder(request.Body).Decode(&p)
	collection := client.Database("API").Collection("Participants")
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	result, _ := collection.InsertOne(ctx, p)
	json.NewEncoder(response).Encode(result)
}

func getParticipants(response http.ResponseWriter, request *http.Request) {
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
		ps = append(ps, p)
	}
	if err := cursor.Err(); err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	json.NewEncoder(response).Encode(ps)
}

func main() {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, _ = mongo.Connect(ctx, clientOptions)
	router := mux.NewRouter()
	router.HandleFunc("/participants", createParticipant).Methods("POST")
	router.HandleFunc("/participants", getParticipants).Methods("GET")
	http.ListenAndServe(":3000", router)
}
