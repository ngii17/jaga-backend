package utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"jaga-backend/internal/config"
)

const geminiAPIURL = "https://generativelanguage.googleapis.com/v1beta/models/gemini-3.5-flash-lite:generateContent"

// GeminiPart bisa berisi teks BIASA, ATAU permintaan panggil fungsi,
// ATAU hasil dari fungsi yang sudah kita jalankan - tidak pernah dua-duanya sekaligus.
type GeminiPart struct {
	Text             string                  `json:"text,omitempty"`
	FunctionCall     *GeminiFunctionCall     `json:"functionCall,omitempty"`
	FunctionResponse *GeminiFunctionResponse `json:"functionResponse,omitempty"`
	ThoughtSignature string                  `json:"thoughtSignature,omitempty"`
}

type GeminiFunctionCall struct {
	Name string                 `json:"name"`
	Args map[string]interface{} `json:"args"`
}

type GeminiFunctionResponse struct {
	Name     string                 `json:"name"`
	Response map[string]interface{} `json:"response"`
}

type GeminiContent struct {
	Role  string       `json:"role"`
	Parts []GeminiPart `json:"parts"`
}

type geminiTool struct {
	FunctionDeclarations []FunctionDeclaration `json:"functionDeclarations"`
}

type geminiRequest struct {
	Contents          []GeminiContent `json:"contents"`
	Tools             []geminiTool    `json:"tools,omitempty"`
	SystemInstruction *GeminiContent  `json:"systemInstruction,omitempty"`
}

type geminiResponse struct {
	Candidates []struct {
		Content GeminiContent `json:"content"`
	} `json:"candidates"`
}

// SendChatToGemini mengirim seluruh riwayat percakapan + system prompt + daftar tombol,
// mengembalikan SATU balasan Gemini (yang bisa berisi teks ATAU permintaan panggil fungsi).
func SendChatToGemini(cfg *config.Config, systemPrompt string, history []GeminiContent) (*GeminiContent, error) {
	reqBody := geminiRequest{
		Contents: history,
		Tools: []geminiTool{
			{FunctionDeclarations: GetJagaFunctionDeclarations()},
		},
		SystemInstruction: &GeminiContent{
			Parts: []GeminiPart{{Text: systemPrompt}},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, errors.New("gagal menyiapkan permintaan ke Gemini")
	}

	url := fmt.Sprintf("%s?key=%s", geminiAPIURL, cfg.GeminiAPIKey)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, errors.New("gagal menghubungi Gemini")
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.New("gagal membaca balasan Gemini")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("gemini mengembalikan error: %s", string(bodyBytes))
	}

	var geminiResp geminiResponse
	if err := json.Unmarshal(bodyBytes, &geminiResp); err != nil {
		return nil, errors.New("gagal membaca format balasan Gemini")
	}

	if len(geminiResp.Candidates) == 0 {
		return nil, errors.New("gemini tidak memberikan balasan")
	}

	return &geminiResp.Candidates[0].Content, nil
}
