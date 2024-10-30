package messages

import (
	"context"
	"fmt"
	"log/slog"
	"web_auth/internal/models"
)

type MessageService struct {
	log         *slog.Logger
	msgProvider MessageProvider
}

type MessageProvider interface {
	GetUserMessages(ctx context.Context, userID int64, limit, offset int) ([]models.Message, error)
}

func New(log *slog.Logger,
	messageProvider MessageProvider,
) *MessageService {
	return &MessageService{
		msgProvider: messageProvider,
		log:         log,
	}
}

func (a *MessageService) GetUserMessages(ctx context.Context, userID int64, limit, offset int) ([]models.Message, error) {
	const op = "auth.GetUserMessages"

	log := a.log.With(slog.String("op", op))
	log.Info("get user messages attempt", slog.Int64("userID", userID))

	messages, err := a.msgProvider.GetUserMessages(ctx, userID, limit, offset)
	if err != nil {
		log.Error("failed to get user messages", err)
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("user messages retrieved successfully", slog.Int("count", len(messages)))
	return messages, nil
}
