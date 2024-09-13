package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"socialmediaapp/src/models"
	"time"

	"cloud.google.com/go/firestore"
)

func AddPost(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Parse the request body into a new Post struct
	var newPost models.Post

	err := json.NewDecoder(r.Body).Decode(&newPost)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Get the user ID (Assuming the user ID is part of the request)
	userId := newPost.UserId

	// Generate a unique Post ID
	newPost.PostId = fmt.Sprintf("%s-%d", userId, time.Now().UnixNano())
	newPost.CreatedAt = time.Now()
	newPost.LastUpdatedAt = newPost.CreatedAt

	// Get the user's Firestore document reference
	userDocRef := usersCollection.Doc(userId)

	// Add the new post to the user's "posts" array
	_, err = userDocRef.Update(context.Background(), []firestore.Update{
		{
			Path:  "Posts",
			Value: firestore.ArrayUnion(newPost),
		},
	})
	if err != nil {
		http.Error(w, "Failed to add post to Firestore: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Respond with the new post object
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newPost)
}

func DelPost() {

}

func UpdatePost() {

}
