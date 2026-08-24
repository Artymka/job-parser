package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/artymka/jobparser/services/telegram-scrapper/internal/models"
)

func (s *Storage) GetChannels() ([]models.Channel, error) {
	const op = "storage_get_channels"
	ctx, cancel := context.WithTimeout(context.Background(), s.requestTimeoutSeconds*time.Second)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, `SELECT id, username, last_message_id FROM channels`)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	res := make([]models.Channel, 0)
	for rows.Next() {
		res = append(res, models.Channel{})
		i := len(res) - 1
		rows.Scan(&res[i].ID, &res[i].Username, &res[i].LastMessageID)
	}

	return res, nil
}

func (s *Storage) SaveLastMessageIDs(channels []models.Channel) error {
	const op = "storage_save_last_message_ids"
	ctx, cancel := context.WithTimeout(context.Background(), s.requestTimeoutSeconds*time.Second)
	defer cancel()

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer tx.Rollback()

	for _, channel := range channels {
		_, err = tx.ExecContext(ctx, `UPDATE channels SET last_message_id=$1 WHERE id=$2`, channel.LastMessageID, channel.ID)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
	}

	tx.Commit()
	return nil
}
