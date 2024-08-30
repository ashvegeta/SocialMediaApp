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

	// add to notifications
	fromUserData := fromUser.Data()
	toUserData := toUser.Data()

	if notfs, ok := toUserData["Notifications"].([]interface{}); ok {
		toUserData["Notifications"] = append(notfs, models.Notification{
			NID:       fmt.Sprintf(ConnReq.To+"_%d", time.Now().UTC().UnixMilli()),
			IsRead:    false,
			TimeStamp: time.Now().UTC().UnixMilli(),
			CType:     "connection",
			MetaData: map[string]string{
				"From":     ConnReq.From,
				"UserName": fromUserData["UserName"].(string),
			},
		})

		_, err := usersCollection.Doc(ConnReq.To).Update(ctx, []firestore.Update{{
			Path:  "Notifications",
			Value: toUserData["Notifications"],
		},
		})
		if err != nil {
			http.Error(w, "500 : Error While Updating Notifications: "+err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		fmt.Println("Error: 'Notifications' field is not of type []interface{}")
	}

	w.Write([]byte("Added Conn Req to Notifications Successfully"))
}

// If User decides to accept the connection request
func AddConnection(w http.ResponseWriter, r *http.Request) {
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

	// check if userId exists
	if ConnReq.From == "" || ConnReq.To == "" {
		http.Error(w, "field userID is empty", http.StatusBadRequest)
		return
	}

	// check if users exists in DB and id is valid
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

	data1 := fromUser.Data()
	data2 := toUser.Data()

	if friends, ok := data1["Friends"].([]interface{}); ok {
		data1["Friends"] = append(friends, ConnReq.To)

		_, err = usersCollection.Doc(ConnReq.From).Update(ctx, []firestore.Update{{
			Path:  "Friends",
			Value: data1["Friends"],
		},
		})
		if err != nil {
			http.Error(w, "Error While Updating Connection: "+err.Error(), http.StatusInternalServerError)
			log.Printf("%v", err.Error())
			return
		}
	} else {
		fmt.Println("Error: 'Friends' field is not of type []interface{}")
	}

	if friends, ok := data2["Friends"].([]interface{}); ok {
		data2["Friends"] = append(friends, ConnReq.From)

		_, err = usersCollection.Doc(ConnReq.To).Update(ctx, []firestore.Update{{
			Path:  "Friends",
			Value: data2["Friends"],
		},
		})
		if err != nil {
			http.Error(w, "Error While Updating Connection: "+err.Error(), http.StatusInternalServerError)
			log.Printf("%v", err.Error())
			return
		}
	} else {
		fmt.Println("Error: 'Friends' field is not of type []interface{}")
	}

	w.Write([]byte("Added User Connection Successfully"))
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
			friends = removeElement(friends, ConnReq.To)
			if len(friends) == 0 {
				fromUserData["Friends"] = []string{}
			} else {
				fromUserData["Friends"] = removeElement(friends, ConnReq.To)
			}
		} else {
			return fmt.Errorf("invalid Friends field type for From user")
		}

		// Remove connection from "To" user
		toUserData := toUser.Data()
		if friends, ok := toUserData["Friends"].([]interface{}); ok {
			friends = removeElement(friends, ConnReq.From)
			if len(friends) == 0 {
				toUserData["Friends"] = []string{}
			} else {
				toUserData["Friends"] = removeElement(friends, ConnReq.From)
			}
		} else {
			return fmt.Errorf("invalid Friends field type for To user")
		}

		// Update both users in Firestore
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
func removeElement(slice []interface{}, elem string) []interface{} {
	var result []interface{}
	for _, v := range slice {
		if v != elem {
			result = append(result, v.(string))
		}
	}
	return result
}
