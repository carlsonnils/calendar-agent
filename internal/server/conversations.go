package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"fake.com/nilspcarlson/internal/dal"
)

// return all conversation information for each conversation
func ListConversationsHandler(w http.ResponseWriter, r *http.Request) {
	// query database for conversations
	conversations, err := dal.ListConversations(context.Background())
	if err != nil {
		fmt.Println("ListConversationsHandler query:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// marshal the conversations into []byte
	conversationsBytes, err := json.Marshal(conversations)
	if err != nil {
		fmt.Println("ListConversationsHandler marshal conversations:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// respond with the conversations in the body
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(conversationsBytes)
}
