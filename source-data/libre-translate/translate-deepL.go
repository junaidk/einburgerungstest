package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

// QuizQuestion represents a single question in the quiz
type QuizQuestion struct {
	ID            int      `json:"id"`
	Question      string   `json:"question"`
	QuestionTrans string   `json:"question_trans"`
	Options       []string `json:"options"`
	OptionsTrans  []string `json:"options_trans"`
	CorrectAnswer int      `json:"correctAnswer"`
}

// Quiz represents the entire quiz data structure
type Quiz struct {
	Title     string         `json:"title"`
	Questions []QuizQuestion `json:"questions"`
}

// DeepLRequest represents the request to the DeepL API
type DeepLRequest struct {
	Text       []string `json:"text"`
	TargetLang string   `json:"target_lang"`
	SourceLang string   `json:"source_lang"`
}

// DeepLResponse represents the response from the DeepL API
type DeepLResponse struct {
	Translations []struct {
		DetectedSourceLanguage string `json:"detected_source_language"`
		Text                   string `json:"text"`
	} `json:"translations"`
}

// Translate text using DeepL API
func translateWithDeepL(texts []string, apiKey string) ([]string, error) {
	// Prepare the request
	requestData := DeepLRequest{
		Text:       texts,
		TargetLang: "EN",
		SourceLang: "DE",
	}

	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %v", err)
	}

	// Create and send the HTTP request
	req, err := http.NewRequest("POST", "https://api-free.deepl.com/v2/translate", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "DeepL-Auth-Key "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	// Read and parse the response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %s", string(body))
	}

	var response DeepLResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %v", err)
	}

	// Extract translated texts
	translations := make([]string, len(response.Translations))
	for i, translation := range response.Translations {
		translations[i] = translation.Text
	}

	return translations, nil
}

func main() {
	// Check if DeepL API key is provided
	apiKey := os.Getenv("DEEPL_API_KEY")
	if apiKey == "" {
		fmt.Println("Error: DEEPL_API_KEY environment variable is required")
		os.Exit(1)
	}

	// Check if input file is provided
	if len(os.Args) < 2 {
		fmt.Println("Error: Input file path is required")
		fmt.Println("Usage: go run main.go <input_file>")
		os.Exit(1)
	}

	// Read JSON from file
	inputFile := os.Args[1]
	inputData, err := ioutil.ReadFile(inputFile)
	if err != nil {
		fmt.Printf("Error reading input file: %v\n", err)
		os.Exit(1)
	}

	// Parse the JSON
	var quiz Quiz
	err = json.Unmarshal(inputData, &quiz)
	if err != nil {
		fmt.Printf("Error parsing JSON: %v\n", err)
		os.Exit(1)
	}

	// Process each question
	for i, question := range quiz.Questions {
		fmt.Println("procession question %d", i)
		// Collect all texts to be translated for this question
		textsToTranslate := make([]string, 0)
		textsToTranslate = append(textsToTranslate, question.Question)
		textsToTranslate = append(textsToTranslate, question.Options...)

		// Translate all texts at once
		translations, err := translateWithDeepL(textsToTranslate, apiKey)
		if err != nil {
			fmt.Printf("Error translating question %d: %v\n", question.ID, err)
			os.Exit(1)
		}

		// Update the question with translations
		quiz.Questions[i].QuestionTrans = translations[0]

		// Update options translations
		optionsTrans := translations[1:] // Skip the first translation (which is the question)
		quiz.Questions[i].OptionsTrans = optionsTrans
		// add simple retelimt
		time.Sleep(1 * time.Second)
	}

	// Output the updated JSON
	outputData, err := json.MarshalIndent(quiz, "", "  ")
	if err != nil {
		fmt.Printf("Error generating output JSON: %v\n", err)
		os.Exit(1)
	}

	// Create output filename based on input filename
	outputFile := inputFile
	if len(os.Args) > 2 {
		outputFile = os.Args[2]
	} else {
		// Add "_translated" suffix to the input filename
		ext := ".json"
		base := inputFile
		if idx := len(inputFile) - len(ext); idx > 0 && inputFile[idx:] == ext {
			base = inputFile[:idx]
		}
		outputFile = base + "_translated" + ext
	}

	// Write to output file
	err = ioutil.WriteFile(outputFile, outputData, 0644)
	if err != nil {
		fmt.Printf("Error writing output file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Translation completed successfully. Output written to %s\n", outputFile)
}
