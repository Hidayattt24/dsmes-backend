package ai_chat

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/dsmes/dsmes-backend/internal/config"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type AIChatService interface {
	SendMessage(ctx context.Context, patientID uuid.UUID, req SendMessageRequest) (*SendMessageResponse, error)
	GetConversations(ctx context.Context, patientID uuid.UUID) ([]ConversationResponse, error)
	CreateConversation(ctx context.Context, patientID uuid.UUID, title string) (*ConversationResponse, error)
	GetMessages(ctx context.Context, patientID uuid.UUID, conversationID uuid.UUID) ([]MessageResponse, error)
	DeleteConversation(ctx context.Context, patientID uuid.UUID, conversationID uuid.UUID) error
}

type aiChatService struct {
	repo          AIChatRepository
	provider      AIProvider
	promptBuilder *PromptBuilder
	config        config.AIConfig
	logger        *zap.Logger
}

func NewAIChatService(repo AIChatRepository, cfg config.AIConfig, logger *zap.Logger) AIChatService {
	return &aiChatService{
		repo:          repo,
		provider:      NewAIProvider(cfg, logger),
		promptBuilder: NewPromptBuilder(),
		config:        cfg,
		logger:        logger,
	}
}

func (s *aiChatService) SendMessage(ctx context.Context, patientID uuid.UUID, req SendMessageRequest) (*SendMessageResponse, error) {
	if strings.TrimSpace(req.Message) == "" {
		return nil, errors.New("message text cannot be empty")
	}

	var convID uuid.UUID
	var err error

	// 1. Resolve or Create Conversation
	if strings.TrimSpace(req.ConversationID) != "" {
		convID, err = uuid.Parse(req.ConversationID)
		if err != nil {
			return nil, fmt.Errorf("invalid conversation_id UUID: %w", err)
		}
		// Security check: verify ownership
		existing, err := s.repo.GetConversationByID(ctx, convID, patientID)
		if err != nil || existing == nil {
			return nil, errors.New("conversation not found or access denied")
		}
	} else {
		// Auto create new conversation with title derived from user message
		title := strings.TrimSpace(req.Message)
		if len(title) > 35 {
			title = title[:35] + "..."
		}
		newConv := &AIConversation{
			ID:        uuid.New(),
			PatientID: patientID,
			Title:     title,
		}
		if err := s.repo.CreateConversation(ctx, newConv); err != nil {
			return nil, fmt.Errorf("failed to create new conversation: %w", err)
		}
		convID = newConv.ID
	}

	// 2. Fetch context memory (last 10 messages)
	history, err := s.repo.GetMessagesByConversation(ctx, convID, patientID, 10)
	if err != nil {
		s.logger.Warn("Failed to fetch context memory messages", zap.Error(err))
	}

	// 3. Save User Message to DB
	userMsg := &AIMessage{
		ID:             uuid.New(),
		ConversationID: convID,
		PatientID:      patientID,
		Role:           "user",
		Message:        strings.TrimSpace(req.Message),
		CreatedAt:      time.Now(),
	}
	if err := s.repo.CreateMessage(ctx, userMsg); err != nil {
		return nil, fmt.Errorf("failed to save user message: %w", err)
	}

	// 4. Retrieve Patient Health Context
	healthCtx, err := s.repo.GetPatientHealthContext(ctx, patientID)
	if err != nil {
		s.logger.Warn("Failed to retrieve patient health context", zap.Error(err))
	}

	// 5. Build System Prompt
	systemPrompt := s.promptBuilder.BuildSystemPrompt(healthCtx)

	// 6. Invoke AI Provider with timing log
	startTime := time.Now()
	assistantText, err := s.provider.GenerateResponse(ctx, systemPrompt, history, req.Message)
	execTime := time.Since(startTime).Milliseconds()

	status := "success"
	errMsg := ""
	if err != nil {
		status = "error"
		errMsg = err.Error()
		s.logger.Error("AI Provider failed to generate response", zap.Error(err))
		assistantText = generateFallbackAnswer(req.Message)
	}

	// Sanitize output: remove any markdown symbols so clean plain text is stored and sent to client
	assistantText = SanitizeAIResponse(assistantText)

	// 7. Audit Log prompt generation
	_ = s.repo.CreatePromptLog(ctx, &AIPromptLog{
		ID:              uuid.New(),
		PatientID:       patientID,
		ConversationID:  &convID,
		GeneratedPrompt: systemPrompt,
		Model:           s.config.Model,
		ExecutionTimeMS: int(execTime),
		Status:          status,
		ErrorMessage:    errMsg,
	})

	// 8. Save Assistant Message to DB
	assistantMsg := &AIMessage{
		ID:             uuid.New(),
		ConversationID: convID,
		PatientID:      patientID,
		Role:           "assistant",
		Message:        assistantText,
		CreatedAt:      time.Now(),
	}
	if err := s.repo.CreateMessage(ctx, assistantMsg); err != nil {
		s.logger.Error("Failed to save assistant message", zap.Error(err))
	}

	// Update conversation updated_at timestamp
	_ = s.repo.UpdateConversationTitle(ctx, convID, "")

	return &SendMessageResponse{
		AssistantMessage: assistantText,
		ConversationID:   convID.String(),
		MessageID:        assistantMsg.ID.String(),
		Timestamp:        assistantMsg.CreatedAt.Format(time.RFC3339),
	}, nil
}

func (s *aiChatService) GetConversations(ctx context.Context, patientID uuid.UUID) ([]ConversationResponse, error) {
	convs, err := s.repo.ListConversationsByPatient(ctx, patientID)
	if err != nil {
		return nil, err
	}

	result := make([]ConversationResponse, 0, len(convs))
	for _, c := range convs {
		result = append(result, ConversationResponse{
			ID:        c.ID.String(),
			PatientID: c.PatientID.String(),
			Title:     c.Title,
			CreatedAt: c.CreatedAt,
			UpdatedAt: c.UpdatedAt,
		})
	}
	return result, nil
}

func (s *aiChatService) CreateConversation(ctx context.Context, patientID uuid.UUID, title string) (*ConversationResponse, error) {
	if strings.TrimSpace(title) == "" {
		title = "Percakapan Baru"
	}
	conv := &AIConversation{
		ID:        uuid.New(),
		PatientID: patientID,
		Title:     title,
	}
	if err := s.repo.CreateConversation(ctx, conv); err != nil {
		return nil, err
	}

	return &ConversationResponse{
		ID:        conv.ID.String(),
		PatientID: conv.PatientID.String(),
		Title:     conv.Title,
		CreatedAt: conv.CreatedAt,
		UpdatedAt: conv.UpdatedAt,
	}, nil
}

func (s *aiChatService) GetMessages(ctx context.Context, patientID uuid.UUID, conversationID uuid.UUID) ([]MessageResponse, error) {
	// Verify conversation ownership
	conv, err := s.repo.GetConversationByID(ctx, conversationID, patientID)
	if err != nil || conv == nil {
		return nil, errors.New("conversation not found or access denied")
	}

	msgs, err := s.repo.GetMessagesByConversation(ctx, conversationID, patientID, 100)
	if err != nil {
		return nil, err
	}

	result := make([]MessageResponse, 0, len(msgs))
	for _, m := range msgs {
		result = append(result, MessageResponse{
			ID:             m.ID.String(),
			ConversationID: m.ConversationID.String(),
			Role:           m.Role,
			Message:        m.Message,
			CreatedAt:      m.CreatedAt,
		})
	}
	return result, nil
}

func (s *aiChatService) DeleteConversation(ctx context.Context, patientID uuid.UUID, conversationID uuid.UUID) error {
	return s.repo.DeleteConversation(ctx, conversationID, patientID)
}
