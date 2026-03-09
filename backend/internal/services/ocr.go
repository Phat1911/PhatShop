package services

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

// ExtractedReceipt holds structured data pulled from a bank transfer receipt image.
type ExtractedReceipt struct {
	TransactionID *string `json:"transaction_id"`
	Amount        *int64  `json:"amount"`
	ReceiverInfo  *string `json:"receiver_info"` // account number or name
	TransferNote  *string `json:"transfer_note"`
	IsSuspicious  bool    `json:"is_suspicious"`
	RawResponse   string  `json:"-"`
}

type OCRService struct {
	apiKey string
}

func NewOCRService(apiKey string) *OCRService {
	return &OCRService{apiKey: apiKey}
}

func (s *OCRService) IsConfigured() bool {
	return s.apiKey != "" && s.apiKey != "your_gemini_api_key_here"
}

// ExtractFromImageFile reads the image at filePath and calls Gemini Vision to extract receipt data.
func (s *OCRService) ExtractFromImageFile(filePath string) (*ExtractedReceipt, error) {
	if !s.IsConfigured() {
		return nil, fmt.Errorf("Gemini API key not configured")
	}

	imgBytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read image: %w", err)
	}

	mimeType := detectMIME(imgBytes)
	b64 := base64.StdEncoding.EncodeToString(imgBytes)

	prompt := `You are a Vietnamese banking receipt analyser. Extract the following from this bank transfer receipt/screenshot and respond with valid JSON only — no markdown, no explanation.

{
  "transaction_id": "<reference/transaction code, e.g. FT24123456789 or similar — null if not found>",
  "amount": <transfer amount as integer in VND — null if not found>,
  "receiver_info": "<receiver account number or full name — null if not found>",
  "transfer_note": "<transfer description/note/content — null if not found>",
  "is_suspicious": <true if the image appears edited, cropped suspiciously, has inconsistent fonts, or looks manipulated — otherwise false>
}`

	reqBody := map[string]any{
		"contents": []map[string]any{
			{
				"parts": []map[string]any{
					{"text": prompt},
					{"inline_data": map[string]any{
						"mime_type": mimeType,
						"data":      b64,
					}},
				},
			},
		},
		"generationConfig": map[string]any{
			"temperature": 0,
		},
	}

	body, _ := json.Marshal(reqBody)
	url := "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash:generateContent?key=" + s.apiKey

	resp, err := http.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("gemini request failed: %w", err)
	}
	defer resp.Body.Close()

	respBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("gemini error %d: %s", resp.StatusCode, string(respBytes))
	}

	// Parse Gemini response envelope
	var geminiResp struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}
	if err := json.Unmarshal(respBytes, &geminiResp); err != nil {
		return nil, fmt.Errorf("failed to parse gemini response: %w", err)
	}
	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("gemini returned empty response")
	}

	rawText := geminiResp.Candidates[0].Content.Parts[0].Text

	// Parse the JSON the model produced
	var result ExtractedReceipt
	if err := json.Unmarshal([]byte(rawText), &result); err != nil {
		// Try stripping markdown code fences if present
		clean := strings.TrimSpace(rawText)
		clean = strings.TrimPrefix(clean, "```json")
		clean = strings.TrimPrefix(clean, "```")
		clean = strings.TrimSuffix(clean, "```")
		if err2 := json.Unmarshal([]byte(clean), &result); err2 != nil {
			return nil, fmt.Errorf("failed to parse extracted JSON: %w", err2)
		}
	}
	result.RawResponse = rawText
	return &result, nil
}

func detectMIME(data []byte) string {
	if len(data) >= 4 {
		// PNG: 89 50 4E 47
		if data[0] == 0x89 && data[1] == 0x50 {
			return "image/png"
		}
		// JPEG: FF D8 FF
		if data[0] == 0xFF && data[1] == 0xD8 {
			return "image/jpeg"
		}
		// WebP: 52 49 46 46 ... 57 45 42 50
		if data[0] == 0x52 && data[1] == 0x49 {
			return "image/webp"
		}
	}
	return "image/jpeg"
}
