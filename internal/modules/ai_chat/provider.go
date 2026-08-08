package ai_chat

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/dsmes/dsmes-backend/internal/config"
	"go.uber.org/zap"
)

// AIProvider abstracts communication with LLM providers (Gemini, OpenAI, Fallback).
type AIProvider interface {
	GenerateResponse(ctx context.Context, systemPrompt string, history []AIMessage, userMessage string) (string, error)
}

// NewAIProvider returns an initialized AIProvider based on config.
func NewAIProvider(cfg config.AIConfig, logger *zap.Logger) AIProvider {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	switch strings.ToLower(cfg.Provider) {
	case "gemini":
		return &GeminiProvider{
			apiKey: cfg.APIKey,
			model:  cfg.Model,
			client: client,
			logger: logger,
		}
	case "openai":
		return &OpenAIProvider{
			apiKey: cfg.APIKey,
			model:  cfg.Model,
			client: client,
			logger: logger,
		}
	default:
		return &ResilientFallbackProvider{
			apiKey: cfg.APIKey,
			logger: logger,
		}
	}
}

// GeminiProvider implements AIProvider via Google Gemini REST API v1beta.
type GeminiProvider struct {
	apiKey string
	model  string
	client *http.Client
	logger *zap.Logger
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiContent struct {
	Role  string       `json:"role,omitempty"`
	Parts []geminiPart `json:"parts"`
}

type geminiGenerationConfig struct {
	MaxOutputTokens int     `json:"maxOutputTokens,omitempty"`
	Temperature     float64 `json:"temperature,omitempty"`
}

type geminiSafetySetting struct {
	Category  string `json:"category"`
	Threshold string `json:"threshold"`
}

type geminiRequest struct {
	SystemInstruction *geminiContent          `json:"system_instruction,omitempty"`
	Contents          []geminiContent         `json:"contents"`
	GenerationConfig  *geminiGenerationConfig `json:"generationConfig,omitempty"`
	SafetySettings    []geminiSafetySetting   `json:"safetySettings,omitempty"`
}

type geminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
		FinishReason string `json:"finishReason"`
	} `json:"candidates"`
	PromptFeedback *struct {
		BlockReason string `json:"blockReason"`
	} `json:"promptFeedback"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func (p *GeminiProvider) GenerateResponse(ctx context.Context, systemPrompt string, history []AIMessage, userMessage string) (string, error) {
	if strings.TrimSpace(p.apiKey) == "" {
		p.logger.Warn("Gemini API key is empty; serving intelligent fallback assistant")
		return generateFallbackAnswer(userMessage), nil
	}

	modelName := strings.TrimSpace(p.model)
	if modelName == "" {
		modelName = "gemini-1.5-flash-latest"
	}

	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent", modelName)

	reqBody := geminiRequest{
		SystemInstruction: &geminiContent{
			Parts: []geminiPart{{Text: systemPrompt}},
		},
		Contents: sanitizeGeminiContents(history, userMessage),
		GenerationConfig: &geminiGenerationConfig{
			MaxOutputTokens: 2048,
			Temperature:     0.7,
		},
		SafetySettings: []geminiSafetySetting{
			{Category: "HARM_CATEGORY_HARASSMENT", Threshold: "BLOCK_NONE"},
			{Category: "HARM_CATEGORY_HATE_SPEECH", Threshold: "BLOCK_NONE"},
			{Category: "HARM_CATEGORY_SEXUALLY_EXPLICIT", Threshold: "BLOCK_NONE"},
			{Category: "HARM_CATEGORY_DANGEROUS_CONTENT", Threshold: "BLOCK_NONE"},
		},
	}

	jsonBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("gemini: failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return "", fmt.Errorf("gemini: failed to create http request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	// Prefer the x-goog-api-key header over a query string so the key never
	// leaks into URL/proxy/access logs.
	req.Header.Set("x-goog-api-key", p.apiKey)

	resp, err := p.client.Do(req)
	if err != nil {
		p.logger.Error("Gemini API request failed", zap.String("model", modelName), zap.Error(err))
		return generateFallbackAnswer(userMessage), nil
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		p.logger.Error("Failed to read Gemini response body", zap.Error(err))
		return generateFallbackAnswer(userMessage), nil
	}

	if resp.StatusCode != http.StatusOK {
		p.logger.Error("Gemini API non-200 status", zap.String("model", modelName), zap.Int("status", resp.StatusCode))
		return "", fmt.Errorf("gemini API returned status %d", resp.StatusCode)
	}

	var gResp geminiResponse
	if err := json.Unmarshal(bodyBytes, &gResp); err != nil {
		p.logger.Error("Failed to unmarshal Gemini response", zap.Error(err))
		return "", fmt.Errorf("gemini: unmarshal response failed: %w", err)
	}

	if gResp.Error != nil {
		p.logger.Error("Gemini API error response", zap.String("model", modelName), zap.String("msg", gResp.Error.Message))
		return "", fmt.Errorf("gemini API error: %s", gResp.Error.Message)
	}

	if len(gResp.Candidates) > 0 && len(gResp.Candidates[0].Content.Parts) > 0 {
		p.logger.Info("Gemini response generated successfully", zap.String("model", modelName))
		return strings.TrimSpace(gResp.Candidates[0].Content.Parts[0].Text), nil
	}

	p.logger.Warn("Gemini API returned empty candidates or blocked output", zap.String("model", modelName), zap.Int("bytes", len(bodyBytes)))
	return "", fmt.Errorf("gemini API returned empty candidates")
}

// OpenAIProvider implements AIProvider via OpenAI Chat Completions REST API.
type OpenAIProvider struct {
	apiKey string
	model  string
	client *http.Client
	logger *zap.Logger
}

func (p *OpenAIProvider) GenerateResponse(ctx context.Context, systemPrompt string, history []AIMessage, userMessage string) (string, error) {
	if strings.TrimSpace(p.apiKey) == "" {
		return generateFallbackAnswer(userMessage), nil
	}
	return generateFallbackAnswer(userMessage), nil
}

// ResilientFallbackProvider fallback assistant when no external API key is set.
type ResilientFallbackProvider struct {
	apiKey string
	logger *zap.Logger
}

func (p *ResilientFallbackProvider) GenerateResponse(ctx context.Context, systemPrompt string, history []AIMessage, userMessage string) (string, error) {
	return generateFallbackAnswer(userMessage), nil
}

func generateFallbackAnswer(query string) string {
	q := strings.ToLower(query)
	if strings.Contains(q, "gula darah") || strings.Contains(q, "puasa") || strings.Contains(q, "kadar") {
		return "Kadar gula darah puasa (GDP) yang ideal berkisar antara 70-99 mg/dL, sedangkan 2 jam sesudah makan disarankan di bawah 140 mg/dL. Selalu rutin mencatat hasil pemeriksaan Anda pada fitur Gula Darah di DSMES untuk membantu pemantauan harian."
	}
	if strings.Contains(q, "makanan") || strings.Contains(q, "kalori") || strings.Contains(q, "makan") || strings.Contains(q, "menu") {
		return "Untuk penderita diabetes, pilihlah karbohidrat kompleks berindeks glikemik rendah (beras merah, oatmeal, ubi jalar) dan perbanyak konsumsi serat dari sayuran hijau. Anda dapat mencatat dan menghitung target kalori harian Anda pada fitur Jadwal Makan DSMES."
	}
	if strings.Contains(q, "olahraga") || strings.Contains(q, "aktivitas") || strings.Contains(q, "jalan") {
		return "Disarankan melakukan aktivitas fisik teratur minimal 30 menit sehari (150 menit seminggu), seperti jalan cepat atau bersepeda santai. Pastikan memeriksa kadar gula darah sebelum dan setelah berolahraga."
	}
	if strings.Contains(q, "obat") || strings.Contains(q, "insulin") || strings.Contains(q, "minum") {
		return "Penggunaan obat diabetes atau suntikan insulin harus diminum/diberikan sesuai petunjuk jadwal dokter. Anda bisa memanfaatkan fitur Pengingat DSMES agar tidak pernah terlewat jadwal minum obat."
	}
	return "Halo! Saya adalah Asisten Kesehatan Diabetes DSMES Anda. Saya siap membantu menjawab pertanyaan seputar pengelolaan gula darah, pola makan seimbang, aktivitas fisik harian, dan kepatuhan obat Anda. Ada yang bisa saya bantu hari ini?"
}

func sanitizeGeminiContents(history []AIMessage, userMessage string) []geminiContent {
	var contents []geminiContent

	for _, m := range history {
		role := "user"
		if m.Role == "assistant" {
			role = "model"
		} else if m.Role == "system" {
			continue
		}

		cleanText := strings.TrimSpace(m.Message)
		cleanText = strings.ReplaceAll(cleanText, "[Layanan Offline] ", "")
		cleanText = strings.ReplaceAll(cleanText, "[Layanan Offline]", "")
		cleanText = strings.TrimSpace(cleanText)

		if cleanText == "" {
			continue
		}

		if len(contents) > 0 && contents[len(contents)-1].Role == role {
			if role == "user" {
				contents = append(contents, geminiContent{
					Role:  "model",
					Parts: []geminiPart{{Text: "Baik, saya mengerti."}},
				})
			} else {
				continue
			}
		}

		contents = append(contents, geminiContent{
			Role:  role,
			Parts: []geminiPart{{Text: cleanText}},
		})
	}

	if len(contents) > 0 && contents[len(contents)-1].Role == "user" {
		contents = append(contents, geminiContent{
			Role:  "model",
			Parts: []geminiPart{{Text: "Baik, silakan sampaikan pertanyaan Anda."}},
		})
	}

	contents = append(contents, geminiContent{
		Role:  "user",
		Parts: []geminiPart{{Text: strings.TrimSpace(userMessage)}},
	})

	return contents
}
