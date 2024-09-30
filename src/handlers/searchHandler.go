package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"google.golang.org/api/iterator"
)

// Search handles search for users by username
func Search(w http.ResponseWriter, r *http.Request) {
	// Get the username parameter from the query string
	username := r.URL.Query().Get("UserName")
	if username == "" {
		http.Error(w, "Missing 'UserName' query parameter", http.StatusBadRequest)
		return
	}

	// Query Firestore for users matching the given username (partial match)
	var users []map[string]interface{}
	ctx := context.Background()

	// Create a query to find all users whose UserName field contains the provided username
	query := usersCollection.Where("UserName", ">=", strings.ToLower(username)).
		Where("UserName", "<", strings.ToLower(username)+"\uf8ff")

	iter := query.Documents(ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Printf("Error getting documents: %v", err)
			http.Error(w, "Error fetching users", http.StatusInternalServerError)
			return
		}

		// Get the document data
		data := doc.Data()
		visibility, _ := data["Visibility"].(string)

		// Create a result map to include only the necessary fields based on visibility
		result := map[string]interface{}{
			"UID":        doc.Ref.ID,
			"EmailId":    data["EmailId"],
			"UserName":   data["UserName"],
			"Visibility": data["Visibility"],
			"Friends":    data["Friends"],
		}

		if visibility == "public" {
			// Include Posts field for public profiles
			if posts, ok := data["Posts"]; ok {
				result["Posts"] = posts
			}
		}

		// Append the formatted result to the users slice
		users = append(users, result)
	}

	// Return the matching users as a JSON response
	w.Header().Set("Content-Type", "application/json")
	if len(users) == 0 {
		json.NewEncoder(w).Encode([]map[string]interface{}{}) // Return empty array if no users found
	} else {
		json.NewEncoder(w).Encode(users) // Return matching users
	}
}
