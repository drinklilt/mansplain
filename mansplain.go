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

	"github.com/ollama/ollama/api"
)

var (
	manpage  string
	model    string
	endpoint string = "http://localhost:11434/api"
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

	// Add mansplaining prompt
	var prompt string = "Role play as a rude and condescending person who is snarky and mansplains a lot. Explain only the useful parts of this documentation to me please. I can handle it."

	var apiRequest api.GenerateRequest
	apiRequest.Model = model
	apiRequest.Prompt = prompt + "\n" + stdout.String()
	apiRequest.KeepAlive = &api.Duration{Duration: 30_000_000_000} // 30 seconds (in nanoseconds)

	data, err := json.Marshal(apiRequest)
	if err != nil {
		fmt.Println("Error marshalling request:", err)
		return
	}

	// Send to ollama API for mansplaining
	req, err := http.NewRequest("POST", endpoint+"/generate", bytes.NewBuffer([]byte(data)))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)

	// Data is in a steam, so scanner is necessary
	respScanner := bufio.NewScanner(resp.Body)

	for respScanner.Scan() {
		var respData api.GenerateResponse
		err := json.Unmarshal([]byte(respScanner.Text()), &respData)
		if err != nil {
			fmt.Println("Error unmarshalling response:", err)
			return
		}
		// Add a newline after the final response so the shell prompt is not on the same line
		if respData.Done {
			fmt.Println(respData.Response)
		} else {
			fmt.Print(respData.Response)
		}
	}

	if err != nil {
		fmt.Println("Error sending request to ollama API:", err)
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
