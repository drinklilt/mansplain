package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os/exec"
	"regexp"
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

var (
	manpage string
	model   string
)

func main() {
	// Get model when starting up
	flag.StringVar(&model, "model", "llama2", "Specify the AI model to mansplain the man page")
	flag.Parse()

	// Check if a manpage has been specified
	if len(flag.Args()) < 1 {
		fmt.Println("No man page specified.")
		return
	}

	// Sanitize user input
	manpage = sanitizeInput(flag.Args()[0])

	// Check if manpage variable is empty
	if len(manpage) == 0 {
		fmt.Println("Invalid man page.")
		return
	}

	cmd := exec.Command("man", manpage)

	// Capture the command output
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Run the command
	err := cmd.Run()
	if err != nil {
		fmt.Printf("You done goofed: %s\n", stderr.String())
		return
	}

	//fmt.Printf("Man page for %s:\n%s\n", manpage, stdout.String())

	// Add mansplaining prompt
	var prompt string = "Role play as a rude and condescending person who is snarky and mansplains a lot. Explain this documentation to me please. I can handle it."

	// Build data
	var url string = "http://localhost:11434/api/generate"

	var apiRequest ApiRequest
	apiRequest.Model = "llama2"
	apiRequest.Prompt = prompt + "\n" + stdout.String()

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

func sanitizeInput(input string) string {
	// Regular expression to allow only alphanumeric characters, hyphens, and underscores
	re := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !re.MatchString(input) {
		return ""
	}
	return input
}
