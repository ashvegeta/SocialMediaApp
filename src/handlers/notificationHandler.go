package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"socialmediaapp/src/models"
	"time"

	"cloud.google.com/go/firestore"
)

// update notifications
func UpdateNotification(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// Decode request body
	var updateReq map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updateReq); err != nil {
		http.Error(w, "400: Bad Request, "+err.Error(), http.StatusBadRequest)
		return
	}

	// Check for valid NID
	if NID, ok := updateReq["NID"].(string); ok {
		// Check for UserId
		if UserId, ok1 := updateReq["UserId"].(string); ok1 {
			user, err := usersCollection.Doc(UserId).Get(ctx)
			if err != nil {
				http.Error(w, "404: User "+UserId+" Not Found", http.StatusBadRequest)
				return
			}

			// Get the user data
			userData := user.Data()

			// Check if Notifications exist
			if notifications, ok := userData["Notifications"].([]interface{}); ok {
				found := false

				// Iterate through the notifications to find the one with matching NID
				for i, notification := range notifications {
					// Convert the interface{} to map[string]interface{}
					notifData := notification.(map[string]interface{})
					if notifData["NID"] == NID {
						found = true
						changed := false

						// Update the notification fields if they exist in the request body
						if isRead, ok := updateReq["IsRead"].(bool); ok {
							notifData["IsRead"] = isRead
							changed = true
						}
						if content, ok := updateReq["Content"].(string); ok {
							notifData["Content"] = content
							changed = true
						}
						if metaData, ok := updateReq["MetaData"].(map[string]string); ok {
							notifData["MetaData"] = metaData
							changed = true
						}
						if cType, ok := updateReq["CType"].(string); ok {
							notifData["CType"] = cType
							changed = true
						}

						// Update the notification in the slice
						if changed {
							notifData["TimeStamp"] = time.Now().UTC().UnixMilli()
						}
						notifications[i] = notifData

						// Update Firestore document with the modified notification
						_, err = usersCollection.Doc(UserId).Update(ctx, []firestore.Update{
							{
								Path:  "Notifications",
								Value: notifications,
							},
						})

						if err != nil {
							http.Error(w, "Error while updating notification: "+err.Error(), http.StatusInternalServerError)
							log.Printf("%v", err.Error())
							return
						}

						w.Write([]byte("Notification Updated Successfully"))
						return
					}
				}

				// If NID is not found
				if !found {
					http.Error(w, "404 : NID "+NID+" does not exist for user "+UserId, http.StatusNotFound)
					return
				}
			} else {
				http.Error(w, "500 : Notifications is not of type []interface{}", http.StatusInternalServerError)
				return
			}
		} else {
			http.Error(w, "400: Bad Request, missing UserId or Type Error (check K,V pair)", http.StatusBadRequest)
			return
		}
	} else {
		http.Error(w, "400: Bad Request, missing NotificationID or Type Error (check K,V pair)", http.StatusBadRequest)
		return
	}
}

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
