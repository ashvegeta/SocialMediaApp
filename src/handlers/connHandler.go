package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"socialmediaapp/src/models"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Initiate or Request to connect to a User
func RequestConnection(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	var ConnReq models.ConnectionRequest
	if err := json.NewDecoder(r.Body).Decode(&ConnReq); err != nil {
		http.Error(w, "400: Bad Request, "+err.Error(), http.StatusBadRequest)
		return
	}
	if ConnReq.From == "" || ConnReq.To == "" {
		http.Error(w, "400: Bad Request, missing userID", http.StatusBadRequest)
		return
	}

	// Fetch From and To users' data
	fromUser, err1 := usersCollection.Doc(ConnReq.From).Get(ctx)
	toUser, err2 := usersCollection.Doc(ConnReq.To).Get(ctx)

	if status.Code(err1) == codes.NotFound || status.Code(err2) == codes.NotFound {
		var errStr string
		if err1 != nil {
			errStr = ConnReq.From + "; "
		}
		if err2 != nil {
			errStr = errStr + ConnReq.To
		}

		http.Error(w, "404 : User Not Found (While Adding Conn) :"+errStr, http.StatusNotFound)
		return
	}

	// Add to notifications for the recipient (To user)
	fromUserData := fromUser.Data()
	toUserData := toUser.Data()

	if notfs, ok := toUserData["Notifications"].([]interface{}); ok {
		toUserData["Notifications"] = append(notfs, models.Notification{
			NID:       fmt.Sprintf(ConnReq.To+"_%d", time.Now().UTC().UnixMilli()),
			IsRead:    false,
			TimeStamp: time.Now().UTC().UnixMilli(),
			CType:     "connRequest",
			MetaData: map[string]string{
				"From":     ConnReq.From,
				"UserName": fromUserData["UserName"].(string),
			},
		})

		_, err := usersCollection.Doc(ConnReq.To).Update(ctx, []firestore.Update{{
			Path:  "Notifications",
			Value: toUserData["Notifications"],
		}})
		if err != nil {
			http.Error(w, "500 : Error While Updating Notifications: "+err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		fmt.Println("Error: 'Notifications' field is not of type []interface{}")
	}

	// Add the connection request to the 'Pending' field of the current user (From user)
	pendingRequest := map[string]interface{}{
		"To":        ConnReq.To,
		"TimeStamp": time.Now().UTC().UnixMilli(),
	}

	_, err := usersCollection.Doc(ConnReq.From).Update(ctx, []firestore.Update{
		{
			Path:  "Pending",
			Value: firestore.ArrayUnion(pendingRequest),
		},
	})
	if err != nil {
		http.Error(w, "500 : Error While Updating Pending Field: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Respond with success
	w.Write([]byte("Added Conn Req to Notifications and Pending Successfully"))
}

// If User decides to accept the connection request
func AddConnection(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	body := json.NewDecoder(r.Body)
	if body == nil {
		http.Error(w, "400 : Bad Request, Body is Empty", http.StatusBadRequest)
		return
	}

	var ConnReq models.ConnectionRequest
	err := body.Decode(&ConnReq)
	if err != nil {
		log.Printf("%v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Check if userId exists
	if ConnReq.From == "" || ConnReq.To == "" || ConnReq.NID == "" {
		http.Error(w, "field From/To or NID is empty", http.StatusBadRequest)
		return
	}

	// Check if users exist in DB and id is valid
	fromUser, err1 := usersCollection.Doc(ConnReq.From).Get(ctx)
	toUser, err2 := usersCollection.Doc(ConnReq.To).Get(ctx)

	if status.Code(err1) == codes.NotFound || status.Code(err2) == codes.NotFound {
		var errStr string
		if err1 != nil {
			errStr = ConnReq.From + "; "
		}
		if err2 != nil {
			errStr = errStr + ConnReq.To
		}

		log.Printf("User Not Found : %v", errStr)
		http.Error(w, "User Not Found (While Adding Conn) :"+errStr, http.StatusNotFound)
		return
	}

	fromUserData := fromUser.Data()
	toUserData := toUser.Data()

	// Add to "To" user's friend list
	if friends, ok := fromUserData["Friends"].([]interface{}); ok {
		fromUserData["Friends"] = append(friends, ConnReq.To)

		_, err = usersCollection.Doc(ConnReq.From).Update(ctx, []firestore.Update{{
			Path:  "Friends",
			Value: fromUserData["Friends"],
		}})
		if err != nil {
			http.Error(w, "Error While Updating Connection: "+err.Error(), http.StatusInternalServerError)
			log.Printf("%v", err.Error())
			return
		}
	} else {
		fmt.Println("Error: 'Friends' field is not of type []interface{}")
	}

	// Add to "From" user's friend list
	if friends, ok := toUserData["Friends"].([]interface{}); ok {
		toUserData["Friends"] = append(friends, ConnReq.From)

		_, err = usersCollection.Doc(ConnReq.To).Update(ctx, []firestore.Update{{
			Path:  "Friends",
			Value: toUserData["Friends"],
		}})
		if err != nil {
			http.Error(w, "Error While Updating Connection: "+err.Error(), http.StatusInternalServerError)
			log.Printf("%v", err.Error())
			return
		}
	} else {
		fmt.Println("Error: 'Friends' field is not of type []interface{}")
	}

	// Send a notification to the "From" user that their connection request was accepted
	if notfs, ok := fromUserData["Notifications"].([]interface{}); ok {
		fromUserData["Notifications"] = append(notfs, models.Notification{
			NID:       fmt.Sprintf(ConnReq.From+"_%d", time.Now().UTC().UnixMilli()),
			IsRead:    false,
			TimeStamp: time.Now().UTC().UnixMilli(),
			Content:   "You are now connected with " + toUserData["UserName"].(string),
			CType:     "connAccepted",
			MetaData: map[string]string{
				"To":       ConnReq.To,
				"UserName": toUserData["UserName"].(string),
			},
		})

		_, err := usersCollection.Doc(ConnReq.From).Update(ctx, []firestore.Update{{
			Path:  "Notifications",
			Value: fromUserData["Notifications"],
		}})
		if err != nil {
			http.Error(w, "500 : Error While Updating Notifications: "+err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		fmt.Println("Error: 'Notifications' field is not of type []interface{}")
	}

	// Remove the pending request from the user's (From) 'Pending' field
	pendingRequests, ok := fromUserData["Pending"].([]interface{})
	if !ok {
		http.Error(w, "500 : Error While Removing Pending Field: ", http.StatusInternalServerError)
		return
	}

	// Find the pending request that matches the "To" user
	var pendingRequestToRemove map[string]interface{}
	for _, pending := range pendingRequests {
		pendingMap, ok := pending.(map[string]interface{})
		if ok && pendingMap["To"] == ConnReq.To {
			pendingRequestToRemove = pendingMap
			break
		}
	}

	if len(pendingRequestToRemove) > 0 {
		// Remove the pending request from the user's (From) 'Pending' field
		_, err = usersCollection.Doc(ConnReq.From).Update(ctx, []firestore.Update{
			{
				Path:  "Pending",
				Value: firestore.ArrayRemove(pendingRequestToRemove),
			},
		})
		if err != nil {
			http.Error(w, "500 : Error While Removing Pending Field: ", http.StatusInternalServerError)
			return
		}
	}

	// Update the "To" user's notification
	if notfs, ok := toUserData["Notifications"].([]interface{}); ok {
		for i, notf := range notfs {
			notificationMap := notf.(map[string]interface{})
			if notificationMap["NID"] == ConnReq.NID {
				// Update the existing notification
				notificationMap["IsRead"] = true
				notificationMap["TimeStamp"] = time.Now().UTC().UnixMilli()
				notificationMap["Content"] = "You have accepted the connection request from " + fromUserData["UserName"].(string)
				notificationMap["CType"] = "connAccepted"
				notfs[i] = notificationMap // Update the list with the modified notification
				break
			}
		}

		_, err = usersCollection.Doc(ConnReq.To).Update(ctx, []firestore.Update{{
			Path:  "Notifications",
			Value: notfs,
		}})
		if err != nil {
			http.Error(w, "500 : Error While Updating To User's Notifications: "+err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		fmt.Println("Error: 'Notifications' field is not of type []interface{} for To User")
	}

	w.Write([]byte("Added User Connection and Updated Notifications Successfully"))
}

// User wants to delete connection
func DelConnection(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// Decode request body
	var ConnReq models.ConnectionRequest
	if err := json.NewDecoder(r.Body).Decode(&ConnReq); err != nil {
		http.Error(w, "400: Bad Request, "+err.Error(), http.StatusBadRequest)
		return
	}
	if ConnReq.From == "" || ConnReq.To == "" {
		http.Error(w, "400: Bad Request, missing userID", http.StatusBadRequest)
		return
	}

	// Execute Firestore Transaction
	err := fStoreClient.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		// Fetch "From" user data
		fromDocRef := usersCollection.Doc(ConnReq.From)
		fromUser, err := tx.Get(fromDocRef)
		if err != nil {
			if status.Code(err) == codes.NotFound {
				return fmt.Errorf("user not found: %s", ConnReq.From)
			}
			return err
		}

		// Fetch "To" user data
		toDocRef := usersCollection.Doc(ConnReq.To)
		toUser, err := tx.Get(toDocRef)
		if err != nil {
			if status.Code(err) == codes.NotFound {
				return fmt.Errorf("user not found: %s", ConnReq.To)
			}
			return err
		}

		// Remove connection from "From" user
		fromUserData := fromUser.Data()
		if friends, ok := fromUserData["Friends"].([]interface{}); ok {
			friends = removeFriends(friends, ConnReq.To)
			if len(friends) == 0 {
				fromUserData["Friends"] = []string{}
			} else {
				fromUserData["Friends"] = friends
			}
		} else {
			return fmt.Errorf("invalid Friends field type for From user")
		}

		// Remove the pending request from the user's (From) 'Pending' field
		pendingRequests, ok := fromUserData["Pending"].([]interface{})
		if !ok {
			http.Error(w, "500 : Error While Removing Pending Field: ", http.StatusInternalServerError)
			return fmt.Errorf("error while removing pending field")
		}

		var updatedPendingList []map[string]interface{}
		for _, pending := range pendingRequests {
			pendingMap, ok := pending.(map[string]interface{})
			if ok && pendingMap["To"] == ConnReq.To {
				continue
			} else {
				updatedPendingList = append(updatedPendingList, pendingMap)
			}

		}

		if len(fromUserData["Pending"].([]interface{})) != len(updatedPendingList) {
			// update the from user's data
			fromUserData["Pending"] = updatedPendingList

			if len(updatedPendingList) == 0 {
				fromUserData["Pending"] = []interface{}{}
			}

			//update From user's pending list
			if err := tx.Set(fromDocRef, map[string]interface{}{"Pending": fromUserData["Pending"]}, firestore.MergeAll); err != nil {
				return err
			}
		}

		// Remove connection from "To" user
		toUserData := toUser.Data()
		if friends, ok := toUserData["Friends"].([]interface{}); ok {
			friends = removeFriends(friends, ConnReq.From)
			if len(friends) == 0 {
				toUserData["Friends"] = []string{}
			} else {
				toUserData["Friends"] = friends
			}
		} else {
			return fmt.Errorf("invalid Friends field type for To user")
		}

		// Update both users' Friends list in Firestore
		if err := tx.Set(fromDocRef, map[string]interface{}{"Friends": fromUserData["Friends"]}, firestore.MergeAll); err != nil {
			return err
		}
		if err := tx.Set(toDocRef, map[string]interface{}{"Friends": toUserData["Friends"]}, firestore.MergeAll); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		http.Error(w, "500: Error Deleting Connection: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte("Connection Deleted Successfully"))
}

// Helper function to remove an element from a slice
func removeFriends(slice []interface{}, elem string) []interface{} {
	var result []interface{}
	for _, v := range slice {
		if v != elem {
			result = append(result, v.(string))
		}
	}
	return result
}
