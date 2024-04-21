package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

type Data struct {
	Model     string    `json:"model"`
	CreatedAt time.Time `json:"created_at"`
	Response  string    `json:"response"`
	Done      bool      `json:"done"`
}

type ApiRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

func main() {
	// Read from stdin
	scanner := bufio.NewScanner(os.Stdin)
	var input string

	for scanner.Scan() {
		input += scanner.Text() + "\n"
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
		return
	}

	// Add mansplaining prompt
	var prompt string = "Mansplain this documentation to me in a snarky an condescending way:"

	// Build data
	var url string = "http://localhost:11434/api/generate"

	var apiRequest ApiRequest
	apiRequest.Model = "llama2"
	apiRequest.Prompt = prompt + "\n" + input

	data, err := json.Marshal(apiRequest)
	if err != nil {
		fmt.Println("Error marshalling request:", err)
		return
	}

	// Send to ollama API for mansplaining
	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(data)))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)

	respScanner := bufio.NewScanner(resp.Body)
	for respScanner.Scan() {
		var respData Data
		err := json.Unmarshal([]byte(respScanner.Text()), &respData)
		if err != nil {
			fmt.Println("Error unmarshalling response:", err)
			return
		}
		fmt.Print(respData.Response)
	}
}
