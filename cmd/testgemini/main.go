package main

import (
	"fmt"
	"log"

	"jaga-backend/internal/config"
	"jaga-backend/internal/utils"
)

func main() {
	cfg := config.Load()

	history := []utils.GeminiContent{
		{
			Role:  "user",
			Parts: []utils.GeminiPart{{Text: "Halo, apa kabar?"}},
		},
	}

	result, err := utils.SendChatToGemini(cfg, "Kamu adalah asisten ramah.", history)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	for _, part := range result.Parts {
		if part.Text != "" {
			fmt.Println("Balasan Gemini:", part.Text)
		}
		if part.FunctionCall != nil {
			fmt.Println("Gemini minta panggil fungsi:", part.FunctionCall.Name)
		}
	}
}
