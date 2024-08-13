package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"socialmediaapp/src/DB"
	"socialmediaapp/src/models"

	"cloud.google.com/go/firestore"
)

var (
	ctx = context.Background()
	// gStoreClient, bucketHandle = DB.GoogleStorage(ctx)
	fStoreClient    = DB.Firestore(ctx)
	usersCollection = fStoreClient.Collection("users")
)

func AddUser(w http.ResponseWriter, r *http.Request) {
	body := json.NewDecoder(r.Body)

	if body == nil {
		http.Error(w, "400 : Bad Request, Body is Empty", http.StatusBadRequest)
		return
	}

	var user models.User
	err := body.Decode(&user)
	if err != nil {
		log.Printf("%v", err)
		http.Error(w, "Error in Decoding Request Body", http.StatusInternalServerError)
		return
	}

	_, err = usersCollection.Doc(user.UserId).Set(context.Background(), map[string]interface{}{
		"uName":   user.UserName,
		"emailId": user.EmailId,
	})

	if err != nil {
		log.Printf("%v", err)
		http.Error(w, "error adding user", http.StatusInternalServerError)
	}
}

func DelUser(userId string) error {
	_, err := usersCollection.Doc(userId).Delete(context.Background())
	if err != nil {
		return fmt.Errorf("error deleting user %v", userId)
	}

	return nil
}

func UpdateUserInfo(user models.User) error {
	// check if userId exists
	if user.UserId == "" {
		return fmt.Errorf("field userID is empty")
	}

	// Get the Value and type of the struct
	val := reflect.ValueOf(user)
	typ := val.Type()

	// declare new map to store fields to be updated
	updatedData := map[string]interface{}{}

	// Iterate over each field in the struct
	for i := 0; i < val.NumField(); i++ {
		field, fieldType := val.Field(i), typ.Field(i)
		if fieldType.Name == "UserId" {
			continue
		}
		updatedData[fieldType.Name] = field.Interface()
	}

	_, err := usersCollection.Doc(user.UserId).Set(context.Background(), updatedData, firestore.MergeAll)
	if err != nil {
		return fmt.Errorf("error updating user %v", user.UserId)
	}

	return nil
}

func AddPost() {

}

func DelPost() {

}

func UpdatePost() {

}

func AddConnection() {

}

func DelConnection() {

}
