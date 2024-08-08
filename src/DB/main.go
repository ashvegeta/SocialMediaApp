package main

import (
	"context"
	"fmt"
	"log"

	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
	// "cloud.google.com/go/firestore"
)

func main() {
	ctx := context.Background()

	// Replace with path to your service account key file
	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	// // declare bucket info
	bucketName := "socialmediaapp-431315.appspot.com"
	bucket := client.Bucket(bucketName)

	// // List objects in the bucket
	it := bucket.Objects(ctx, nil)

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
}
