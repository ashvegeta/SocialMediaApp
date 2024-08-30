package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"reflect"
	"socialmediaapp/src/DB"
	"socialmediaapp/src/models"

	"cloud.google.com/go/firestore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// check if userId exists
	if user.UserId == "" {
		http.Error(w, "field userID is empty", http.StatusBadRequest)
		return
	}

	_, err = usersCollection.Doc(user.UserId).Set(context.Background(), map[string]interface{}{
		"UserName":      user.UserName,
		"EmailId":       user.EmailId,
		"Visibility":    "private",
		"Posts":         []models.Post{},
		"Friends":       []string{},
		"Notifications": []models.Notification{},
		"ChatHistory":   [][]models.Message{},
	})
	if err != nil {
		log.Printf("%v", err)
		http.Error(w, "error adding user :"+err.Error(), http.StatusInternalServerError)
	}

	w.Write([]byte("Added User Successfully"))
}

func DelUser(w http.ResponseWriter, r *http.Request) {
	body := json.NewDecoder(r.Body)
	if body == nil {
		http.Error(w, "400 : Bad Request, Body is Empty", http.StatusBadRequest)
		return
	}

	var user models.User
	err := body.Decode(&user)
	if err != nil {
		log.Printf("%v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// check if userId exists
	if user.UserId == "" {
		http.Error(w, "field userID is empty", http.StatusBadRequest)
		return
	}

	docRef := usersCollection.Doc(user.UserId)

	// Check if the document exists
	docSnap, err := docRef.Get(context.Background())
	if err != nil {
		if status.Code(err) == codes.NotFound {
			// Document does not exist, return 404 Not Found
			http.Error(w, "Document not found", http.StatusNotFound)
			return
		}

		// Other errors
		http.Error(w, "Failed to check document existence during deletion", http.StatusInternalServerError)
		return
	}

	if !docSnap.Exists() {
		// Document does not exist, return 404 Not Found
		http.Error(w, "Document not found", http.StatusNotFound)
		return
	}

	// Proceed with the deletion if the document exists
	_, err = docRef.Delete(ctx)
	if err != nil {
		http.Error(w, "Failed to delete document", http.StatusInternalServerError)
		return
	}

	w.Write([]byte("Deleted User Successfully"))
}

func UpdateUserInfo(w http.ResponseWriter, r *http.Request) {
	body := json.NewDecoder(r.Body)
	if body == nil {
		http.Error(w, "400 : Bad Request, Body is Empty", http.StatusBadRequest)
		return
	}

	var user models.User
	err := body.Decode(&user)
	if err != nil {
		log.Printf("%v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// check if userId exists
	if user.UserId == "" {
		http.Error(w, "field userID is empty", http.StatusBadRequest)
		return
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

	_, err = usersCollection.Doc(user.UserId).Set(context.Background(), updatedData, firestore.MergeAll)
	if err != nil {
		http.Error(w, "error updating user"+user.UserId+": "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte("Updated User successfully"))
}
