package main

import (
	"context"
	"fmt"
	"log"

	"cloud.google.com/go/firestore"
	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
)

func GoogleStorage(ctx context.Context) (*storage.Client, *storage.BucketHandle) {
	// Replace with path to your service account key file
	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// // declare bucket info
	bucketName := "socialmediaapp-431315.appspot.com"
	bucket := client.Bucket(bucketName)

	return client, bucket
}

func Firestore(ctx context.Context) *firestore.Client {
	// Sets your Google Cloud Platform project ID.
	projectID := "socialmediaapp-431315"

	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	return client
}

func main() {
	// Google Storage
	ctx1 := context.Background()
	gsClient, bucket := GoogleStorage(ctx1)
	defer gsClient.Close()

	// List objects in the bucket
	it := bucket.Objects(ctx1, nil)

	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Object: %s\n", attrs.Name)
	}

	// Firestore
	ctx2 := context.Background()

	fsClient := Firestore(ctx2)
	defer fsClient.Close()

	_, _ , err := fsClient.Collection("users").Add(ctx2, map[string]interface{}{
		"first": "Ada",
        "last":  "Lovelace",
        "born":  1815,
	})

	if err != nil {
		log.Fatalf("failed to add entry : %v", err)
	}
}
