package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang_mongo_api/db"
	"log"
	"net/http"
	"time"
)

var client *mongo.Client

type Person struct {
	ID        primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Firstname string             `json:"firstname,omitempty" bson:"firstname,omitempty"`
	Lastname  string             `json:"lastname,omitempty" bson:"lastname,omitempty"`
}

func main() {
	cache := make(map[primitive.ObjectID][2]string)
	client = db.Connect()
	router := mux.NewRouter()
	router.StrictSlash(true)
	go updateCache(cache)

	router.HandleFunc("/", createRecord).Methods(http.MethodPost)
	router.HandleFunc("/", getRecords).Methods(http.MethodGet)
	router.HandleFunc("/{id}/", deleteRecord).Methods(http.MethodDelete)
	router.HandleFunc("/{id}/", updateRecord).Methods(http.MethodPut)

	log.Fatal(http.ListenAndServe("localhost:8080", router))
}

func createRecord(w http.ResponseWriter, req *http.Request) {
	var person Person
	_ = json.NewDecoder(req.Body).Decode(&person)
	collection := client.Database("test").Collection("collection")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	result, _ := collection.InsertOne(ctx, person)
	w.Header().Set("content-type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func getRecords(w http.ResponseWriter, req *http.Request) {
	var results []Person
	collection := client.Database("test").Collection("collection")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	options := options.Find()
	filter := bson.M{}

	cur, err := collection.Find(ctx, filter, options)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}

	for cur.Next(ctx) {
		var elem Person
		err := cur.Decode(&elem)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{ "message": "` + err.Error() + `" }`))
			return
		}
		results = append(results, elem)
	}

	w.Header().Set("content-type", "application/json")
	json.NewEncoder(w).Encode(results)
}

func deleteRecord(w http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)
	id, _ := params["id"]
	collection := client.Database("test").Collection("collection")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	filter := bson.M{"_id": id}
	deleteResult, err := collection.DeleteOne(ctx, filter)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}

	w.Header().Set("content-type", "application/json")
	json.NewEncoder(w).Encode(deleteResult)
}

func updateRecord(w http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)
	id, _ := primitive.ObjectIDFromHex(params["id"])
	var person Person
	_ = json.NewDecoder(req.Body).Decode(&person)
	filter := bson.M{"_id": id}
	collection := client.Database("test").Collection("collection")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	result, err := collection.UpdateOne(
		ctx,
		filter,
		bson.D{
			{"$set", bson.D{{"Firstname", person.Firstname}, {"Lastname", person.Lastname}}},
		},
	)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}

	w.Header().Set("content-type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func updateCache(cache map[primitive.ObjectID][2]string) {
	for {
		collection := client.Database("test").Collection("collection")
		ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
		filter := bson.M{}
		options := options.Find()
		cur, err := collection.Find(ctx, filter, options)
		if err != nil {
			return
		}
		for cur.Next(ctx) {
			var elem Person
			err := cur.Decode(&elem)
			if err != nil {
				return
			}
			cache[elem.ID] = [2]string{elem.Firstname, elem.Lastname}
			fmt.Println("Обновление кеша завершено. Записей всего:", len(cache))
			time.Sleep(time.Second * 100)
		}
	}
}
