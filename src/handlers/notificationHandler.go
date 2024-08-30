package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"socialmediaapp/src/models"

	"cloud.google.com/go/firestore"
)

// delete notifications
func DelNotification(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// Decode request body
	var ConnReq map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&ConnReq); err != nil {
		http.Error(w, "400: Bad Request, "+err.Error(), http.StatusBadRequest)
		return
	}

	if UserId, ok1 := ConnReq["UserId"].(string); ok1 {
		if NID, ok2 := ConnReq["NID"].(string); ok2 {
			user, err := usersCollection.Doc(UserId).Get(ctx)
			if err != nil {
				http.Error(w, "404: User "+UserId+" Not Found", http.StatusBadRequest)
				return
			}

			// Remove notification from slice and update
			userData := user.Data()
			if notifications, ok := userData["Notifications"].([]interface{}); ok {
				beforeLen := len(notifications)
				notifications = removeNotifications(notifications, NID)

				// Sanitation check
				if len(notifications) == beforeLen {
					http.Error(w, "404 : NID "+NID+" does not exist for user "+UserId, http.StatusNotFound)
					return
				} else if len(notifications) == 0 {
					userData["Notifications"] = []models.Notification{}
				} else {
					userData["Notifications"] = notifications
				}

				// Update the new notifications list to firestore
				if _, err = usersCollection.Doc(UserId).Update(ctx, []firestore.Update{{
					Path:  "Notifications",
					Value: userData["Notifications"],
				},
				}); err != nil {
					http.Error(w, "Error While Updating Connection: "+err.Error(), http.StatusInternalServerError)
					log.Printf("%v", err.Error())
					return
				}
			} else {
				http.Error(w, "500 : Notifications is Not of type []Interface{}", http.StatusInternalServerError)
				return
			}

		} else {
			http.Error(w, "400: Bad Request, missing NotificationID or Type Error (check K,V pair)", http.StatusBadRequest)
			return
		}
	} else {
		http.Error(w, "400: Bad Request, missing UserId or Type Error (check K,V pair)", http.StatusBadRequest)
		return
	}

	w.Write([]byte("Notification Deleted Successfully"))
}

func removeNotifications(slice []interface{}, elem string) []interface{} {
	var result []interface{}

	for _, v := range slice {
		v1 := v.(map[string]interface{})

		if val, ok := v1["NID"]; ok && val != elem {
			result = append(result, val)
		}
	}
	return result
}
