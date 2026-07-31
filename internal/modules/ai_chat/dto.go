package ai_chat

import "time"

// SendMessageRequest payload sent from mobile app to POST /api/v1/ai/chat
type SendMessageRequest struct {
	ConversationID string `json:"conversation_id,omitempty"`
	Message        string `json:"message" validate:"required,min=1"`
}

// SendMessageResponse DTO returned to client after assistant generates a response
type SendMessageResponse struct {
	AssistantMessage string `json:"assistant_message"`
	ConversationID   string `json:"conversation_id"`
	MessageID        string `json:"message_id"`
	Timestamp        string `json:"timestamp"`
}

// CreateConversationRequest payload to POST /api/v1/ai/conversations
type CreateConversationRequest struct {
	Title string `json:"title" validate:"required,min=1,max=255"`
}

// ConversationResponse summary item returned in list/get endpoints
type ConversationResponse struct {
	ID        string            `json:"id"`
	PatientID string            `json:"patient_id"`
	Title     string            `json:"title"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
	Messages  []MessageResponse `json:"messages,omitempty"`
}

// MessageResponse item representation inside conversation details
type MessageResponse struct {
	ID             string    `json:"id"`
	ConversationID string    `json:"conversation_id"`
	Role           string    `json:"role"`
	Message        string    `json:"message"`
	CreatedAt      time.Time `json:"created_at"`
}
