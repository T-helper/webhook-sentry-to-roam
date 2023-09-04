package main
    
import (
	"encoding/json"
	"log"
	"net/http"
)

type WebhookPayload struct {
	Event string `json:"event"`
	Data  struct {
		Message string `json:"message"`
	} `json:"data"`
}

func webhookHandler(w http.ResponseWriter, r *http.Request) {
	var payload WebhookPayload
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
   
	log.Printf("Received webhook: Event=%s, Message=%s", payload.Event, payload.Data.Message)
   
	// Perform any necessary actions based on the received payload
   
	w.WriteHeader(http.StatusOK)
}

func main() {
	http.HandleFunc("/webhook", webhookHandler)
   
	log.Println("Webhook system is running on http://localhost:8080/webhook")
   
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}