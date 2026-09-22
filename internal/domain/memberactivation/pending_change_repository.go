package memberactivation

import (
	"context"
	"encoding/json"

	"github.com/jmoiron/sqlx"
	"github.com/oklog/ulid/v2"
)

type PendingChangeRepository interface {
	Create(ctx context.Context, memberID string, changes map[string]interface{}) error
}

type pendingChangeRepository struct {
	db *sqlx.DB
}

func NewPendingChangeRepository(db *sqlx.DB) PendingChangeRepository {
	return &pendingChangeRepository{db: db}
}

func (r *pendingChangeRepository) Create(ctx context.Context, memberID string, changes map[string]interface{}) error {
	data, err := json.Marshal(changes)
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx,
		"INSERT INTO member_pending_changes (id, member_id, changes) VALUES ($1, $2, $3)",
		ulid.Make().String(), memberID, data)
	return err
}
