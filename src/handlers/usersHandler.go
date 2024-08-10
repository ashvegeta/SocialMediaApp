package handlers

import (
	"context"
	"log"
	"socialmediaapp/src/DB"
	"socialmediaapp/src/models"
)

var (
	ctx = context.Background()
	// gStoreClient, bucketHandle = DB.GoogleStorage(ctx)
	fStoreClient = DB.Firestore(ctx)
)

func AddUser(user models.User) {
	_, _, err := fStoreClient.Collection("users").Add(context.Background(), map[string]interface{}{
		"uName": user.UserName,
	})

	if err != nil {
		log.Printf("error adding user %v", user)
		return
	}
}

func DelUser(userId string) {
	_, err := fStoreClient.Collection("users").Doc(userId).Delete(context.Background())
	if err != nil {
		log.Printf("error while deleting user")
		return
	}
}

func UpdateUserInfo() {

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
