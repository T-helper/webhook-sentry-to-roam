package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

type IncomingPayload struct {
	ID              string   `json:"id"`
	URL             string   `json:"url"`
	Level           string   `json:"level"`
	Logger          *string  `json:"logger"`
	Culprit         string   `json:"culprit"`
	Message         string   `json:"message"`
	Project         string   `json:"project"`
	ProjectName     string   `json:"project_name"`
	ProjectSlug     string   `json:"project_slug"`
	TriggeringRules []string `json:"triggering_rules"`
	Event           Event    `json:"Event"`
}

type Event struct {
	Environment string `json:"environment"`
}

type OutgoingPayload struct {
	Recipients []string `json:"recipients"`
	Sender     struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"sender"`
	Text string `json:"text"`
}

func webhookHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Only POST is supported", http.StatusMethodNotAllowed)
		return
	}

	jsonData, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	url := os.Getenv("URL")

	if err := sendPostRequest(jsonData, url); err != nil {
		http.Error(w, "Failed to send POST request", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Successfully processed webhook"))
}

func sendPostRequest(jsonData []byte, url string) error {
	var incomingPayload IncomingPayload
	err := json.Unmarshal(jsonData, &incomingPayload)
	if err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	// Convert the parsed data into a table-like structure in the specified order
	table := fmt.Sprintf(
		"===========================================================\n\nProject: %s\n\nID: %s\n\nLevel: %s\n\nEnvironment: %s\n\nCulprit: %s\n\nMessage: %s\n\nURL: %s\n\n===========================================================",
		incomingPayload.Project,
		incomingPayload.ID,
		incomingPayload.Level,
		incomingPayload.Event.Environment,
		incomingPayload.Culprit,
		incomingPayload.Message,
		incomingPayload.URL,
	)

	// Handle the triggering rules
	if len(incomingPayload.TriggeringRules) > 0 {
		table += "Triggering Rules:\n\n"
		for _, rule := range incomingPayload.TriggeringRules {
			table += "- " + rule + "\n\n"
		}
	}

	outgoingPayload := OutgoingPayload{
		Recipients: strings.Split(os.Getenv("RECIPIENTS"), ","),
		Text:       table,
	}
	outgoingPayload.Sender.ID = "sentry"
	outgoingPayload.Sender.Name = "Sentry"

	payloadBytes, err := json.Marshal(outgoingPayload)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %v", err)
	}

	client := &http.Client{}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	// Add Authorization header
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", os.Getenv("TOKEN")))
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send POST request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned non-200 status code: %d", resp.StatusCode)
	}

	return nil
}

func main() {
	http.HandleFunc("/webhook", webhookHandler)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Printf("Server listening on port %s...\n\n\n\n", port)
	http.ListenAndServe(":"+port, nil)
}
