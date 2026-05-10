package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
)

type ChatRequest struct {
	Message string `json:"message"`
}

type OllamaMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OllamaChatRequest struct {
	Model    string          `json:"model"`
	Messages []OllamaMessage `json:"messages"`
	Stream   bool            `json:"stream"`
}

type OllamaChatResponse struct {
	Message OllamaMessage `json:"message"`
	Done    bool          `json:"done"`
}

type ChatResponse struct {
	Reply string `json:"reply"`
}

func main() {
	port := getEnv("SERVER_PORT", "8080")

	http.HandleFunc("/health", withCORS(healthHandler))
	http.HandleFunc("/api/chat", withCORS(chatHandler))

	log.Printf("Go backend running on :%s", port)

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]string{
		"status": "ok",
	})
}

func chatHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Message == "" {
		http.Error(w, "message is required", http.StatusBadRequest)
		return
	}

	reply, err := callOllama(req.Message)
	if err != nil {
		log.Println("ollama error:", err)
		http.Error(w, "failed to call ollama", http.StatusInternalServerError)
		return
	}

	writeJSON(w, ChatResponse{
		Reply: reply,
	})
}

func callOllama(userMessage string) (string, error) {
	ollamaHost := getEnv("OLLAMA_HOST", "http://ollama:11434")
	model := getEnv("OLLAMA_MODEL", "qwen3:4b")

	payload := OllamaChatRequest{
		Model: model,
		Messages: []OllamaMessage{
			{
				Role:    "system",
				Content: "あなたは日本語で簡潔に回答するアシスタントです。",
			},
			{
				Role:    "user",
				Content: userMessage,
			},
		},
		Stream: false,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	resp, err := http.Post(
		ollamaHost+"/api/chat",
		"application/json",
		bytes.NewReader(body),
	)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		log.Println("ollama response:", string(respBody))
		return "", statusError(resp.Status)
	}

	var ollamaResp OllamaChatResponse
	if err := json.Unmarshal(respBody, &ollamaResp); err != nil {
		return "", err
	}

	return ollamaResp.Message.Content, nil
}

func withCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// React/Vite から叩く想定
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		next(w, r)
	}
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(v); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

type statusError string

func (e statusError) Error() string {
	return string(e)
}
