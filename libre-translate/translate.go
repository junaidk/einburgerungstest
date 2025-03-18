package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

// Input JSON structure
type InputData struct {
	Title     string     `json:"title"`
	Questions []Question `json:"questions"`
}

type Question struct {
	ID       int      `json:"id"`
	Question string   `json:"question"`
	Options  []string `json:"options"`
	//CorrectAnswer string   `json:"correctAnswer"`
}

// Output JSON structure
type OutputData struct {
	Title     string             `json:"title"`
	Questions []EnhancedQuestion `json:"questions"`
}

type EnhancedQuestion struct {
	ID            int      `json:"id"`
	Question      string   `json:"question"`
	QuestionTrans string   `json:"question_trans"`
	Options       []string `json:"options"`
	OptionsTrans  []string `json:"options_trans"`
	CorrectAnswer string   `json:"correctAnswer"`
}

// Translation request and response structures
type TranslationRequest struct {
	Query        string `json:"q"`
	Source       string `json:"source"`
	Target       string `json:"target"`
	Format       string `json:"format"`
	Alternatives int    `json:"alternatives"`
}

type TranslationResponse struct {
	TranslatedText string `json:"translatedText"`
}

// Client for making API calls with retries
type TranslationClient struct {
	client     *http.Client
	apiURL     string
	maxRetries int
	retryDelay time.Duration
}

func NewTranslationClient(apiURL string) *TranslationClient {
	return &TranslationClient{
		client:     &http.Client{Timeout: 10 * time.Second},
		apiURL:     apiURL,
		maxRetries: 3,
		retryDelay: 500 * time.Millisecond,
	}
}

func (tc *TranslationClient) Translate(text, source, target string) (string, error) {
	request := TranslationRequest{
		Query:        text,
		Source:       source,
		Target:       target,
		Format:       "text",
		Alternatives: 1,
	}

	requestBody, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("error marshaling request: %w", err)
	}

	var lastErr error
	for attempt := 0; attempt <= tc.maxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(tc.retryDelay)
			log.Printf("Retrying translation (attempt %d/%d)", attempt, tc.maxRetries)
		}

		req, err := http.NewRequest("POST", tc.apiURL, bytes.NewBuffer(requestBody))
		if err != nil {
			lastErr = err
			continue
		}

		req.Header.Set("Content-Type", "application/json")
		resp, err := tc.client.Do(req)
		if err != nil {
			lastErr = err
			continue
		}

		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			lastErr = fmt.Errorf("API returned non-200 status: %d, body: %s", resp.StatusCode, string(body))
			continue
		}

		var translationResp TranslationResponse
		if err := json.NewDecoder(resp.Body).Decode(&translationResp); err != nil {
			lastErr = err
			continue
		}

		return translationResp.TranslatedText, nil
	}

	return "", fmt.Errorf("translation failed after %d attempts: %w", tc.maxRetries, lastErr)
}

func main() {
	// Define command-line flags
	//inputFile := flag.String("in", "", "Input JSON file path")
	outputFile := flag.String("out", "output.json", "Output JSON file path")
	apiURL := flag.String("api", "http://localhost:5000/translate", "Translation API URL")
	flag.Parse()

	// Validate input file parameter
	// if *inputFile == "" {
	// 	log.Fatal("Input file is required. Use -in flag to specify the input file path.")
	// }

	baseFiles := []string{"german-citizenship-test-part1.json",
		"german-citizenship-test-part2.json",
		"german-citizenship-test-part3.json",
		"german-citizenship-test-part4.json",
		"german-citizenship-test-part5.json",
		"german-citizenship-test-part6.json",
		"german-citizenship-test-berlin.json"}

	var data InputData
	for _, file := range baseFiles {
		inputFile := file
		// Read input file
		fmt.Println("Reading input file: ", inputFile)
		inputData, err := os.ReadFile("../" + inputFile)
		if err != nil {
			log.Fatalf("Error reading input file: %v", err)
		}

		// Parse JSON
		var fileData InputData
		if err := json.Unmarshal(inputData, &fileData); err != nil {
			log.Fatalf("Error parsing JSON: %v", err)
		}
		data.Questions = append(data.Questions, fileData.Questions...)
	}

	fmt.Printf("Read %d questions from input files\n", len(data.Questions))

	// Initialize translation client
	translationClient := NewTranslationClient(*apiURL)

	// Process each question
	outputData := OutputData{
		Title:     data.Title,
		Questions: make([]EnhancedQuestion, 0, len(data.Questions)),
	}

	for _, q := range data.Questions {
		fmt.Printf("Processing question ID: %d\n", q.ID)

		// Translate question
		questionTrans, err := translationClient.Translate(q.Question, "de", "en")
		if err != nil {
			log.Printf("Error translating question %d: %v", q.ID, err)
			questionTrans = "Translation error"
		}

		// Translate options
		optionsTrans := make([]string, 0, len(q.Options))
		for _, option := range q.Options {
			optionTrans, err := translationClient.Translate(option, "de", "en")
			if err != nil {
				log.Printf("Error translating option '%s': %v", option, err)
				optionTrans = "Translation error"
			}
			optionsTrans = append(optionsTrans, optionTrans)
		}

		// Create enhanced question
		enhancedQuestion := EnhancedQuestion{
			ID:            q.ID,
			Question:      q.Question,
			QuestionTrans: questionTrans,
			Options:       q.Options,
			OptionsTrans:  optionsTrans,
			CorrectAnswer: "",
		}

		outputData.Questions = append(outputData.Questions, enhancedQuestion)
	}

	// Convert to JSON
	outputJSON, err := json.MarshalIndent(outputData, "", "    ")
	if err != nil {
		log.Fatalf("Error generating output JSON: %v", err)
	}

	// Write to output file
	if err := os.WriteFile(*outputFile, outputJSON, 0644); err != nil {
		log.Fatalf("Error writing output file: %v", err)
	}

	fmt.Printf("Translation completed successfully. Output written to %s\n", *outputFile)
}
