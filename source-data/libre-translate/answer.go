package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

// Question represents the structure of a question in the input file
type Question struct {
	ID            int      `json:"id"`
	Question      string   `json:"question"`
	QuestionTrans string   `json:"question_trans"`
	Options       []string `json:"options"`
	OptionsTrans  []string `json:"options_trans"`
	CorrectAnswer int      `json:"correctAnswer,omitempty"`
}

// QuestionsInput represents the structure of the input file
type QuestionsInput struct {
	Title     string     `json:"title"`
	Questions []Question `json:"questions"`
}

// AnswerQuestion represents the structure of an answer in the answer file
type AnswerQuestion struct {
	ID            int `json:"id"`
	CorrectAnswer int `json:"correctAnswer"`
}

// AnswersFile represents the structure of the answer file
type AnswersFile struct {
	Questions []AnswerQuestion `json:"questions"`
}

func main() {
	if len(os.Args) != 4 {
		fmt.Println("Usage: go run program.go <input_file> <answers_file> <output_file>")
		os.Exit(1)
	}

	inputFile := os.Args[1]
	//answersFile := os.Args[2]
	outputFile := os.Args[3]

	// Create output directory if it doesn't exist
	outputDir := filepath.Dir(outputFile)
	if outputDir != "." && outputDir != "" {
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			fmt.Printf("Error creating output directory: %v\n", err)
			os.Exit(1)
		}
	}

	// Read and parse the input file
	inputData, err := ioutil.ReadFile(inputFile)
	if err != nil {
		fmt.Printf("Error reading input file: %v\n", err)
		os.Exit(1)
	}

	var questionsInput QuestionsInput
	if err := json.Unmarshal(inputData, &questionsInput); err != nil {
		fmt.Printf("Error parsing input file: %v\n", err)
		os.Exit(1)
	}

	baseFiles := []string{
		"correct-answers-part1.json",
		"correct-answers-part2.json",
		"correct-answers-part3.json",
		"correct-answers-part4.json",
		"correct-answers-part5.json",
		"correct-answers-part6.json",
		"correct-answers-part7.json",
		"correct-answers-berlin.json",
	}

	answerMap := make(map[int]int)

	for _, file := range baseFiles {
		inputFile := file
		// Read input file
		fmt.Println("Reading input file: ", inputFile)
		answersData, err := ioutil.ReadFile("../answer-data/" + inputFile)
		if err != nil {
			fmt.Printf("Error reading input file: %v\n", err)
			os.Exit(1)
		}

		var answers AnswersFile
		if err := json.Unmarshal(answersData, &answers); err != nil {
			fmt.Printf("Error parsing answers file: %v\n", err)
			os.Exit(1)
		}

		for _, answer := range answers.Questions {
			answerMap[answer.ID] = answer.CorrectAnswer
		}
	}

	// Create a map of answers by question ID for easier lookup

	// Update the correct answers in the input questions
	for i, question := range questionsInput.Questions {
		if correctAnswer, exists := answerMap[question.ID]; exists {
			questionsInput.Questions[i].CorrectAnswer = correctAnswer
		}
	}

	// Convert the updated input to JSON
	outputData, err := json.MarshalIndent(questionsInput, "", "  ")
	if err != nil {
		fmt.Printf("Error generating output JSON: %v\n", err)
		os.Exit(1)
	}

	// Write the output to the specified file
	if err := ioutil.WriteFile(outputFile, outputData, 0644); err != nil {
		fmt.Printf("Error writing output file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully populated answers and created %s\n", outputFile)
}
