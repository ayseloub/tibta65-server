package pemilu

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"sort"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"github.com/oklog/ulid/v2"

	"github.com/Tibta65web/tibta65-server/internal/domain/auth"
	appMiddleware "github.com/Tibta65web/tibta65-server/pkg/middleware"
	"github.com/Tibta65web/tibta65-server/pkg/response"
)

var (
	ErrNoVoteToVoid       = errors.New("anggota ini belum memberikan suara atau suaranya sudah dibatalkan")
	ErrNoVoidedVote       = errors.New("suara anggota ini tidak sedang dibatalkan")
	ErrVoidReasonRequired = errors.New("alasan wajib diisi")
	ErrVoidMemberNotFound = errors.New("anggota tidak ditemukan")
	ErrWinnerWouldChange  = errors.New("tindakan ini akan mengubah pemenang pemilu sementara. Pastikan sudah disetujui pengurus sebelum melanjutkan")
)

type VoidLog struct {
	ID           string `db:"id" json:"id"`
	MemberID     string `db:"member_id" json:"member_id"`
	MemberName   string `db:"member_name" json:"member_name"`
	MemberNumber string `db:"member_number" json:"member_number"`
	Action       string `db:"action" json:"action"`
	Reason       string `db:"reason" json:"reason"`
	AdminID      string `db:"admin_id" json:"admin_id"`
	AdminName    string `db:"admin_name" json:"admin_name"`
	CreatedAt    string `db:"created_at" json:"created_at"`
	CanRestore   bool   `db:"can_restore" json:"can_restore"`
}

// Voter cuma berisi identitas pemilih. Pilihan kandidatnya sengaja gak pernah diambil.
type Voter struct {
	ID           string  `db:"id" json:"id"`
	FullName     string  `db:"full_name" json:"full_name"`
	MemberNumber string  `db:"member_number" json:"member_number"`
	KordaName    *string `db:"korda_name" json:"korda_name"`
}

type VoidRepository interface {
	ListLogs(ctx context.Context) ([]VoidLog, error)
	CountVoided(ctx context.Context) (int, error)
	SearchVoters(ctx context.Context, search string) ([]Voter, error)
	SetVoided(ctx context.Context, memberID, adminID, reason string, void, confirmWinnerChange bool) (bool, error)
}

type voidRepository struct {
	db *sqlx.DB
}

func NewVoidRepository(db *sqlx.DB) VoidRepository {
	return &voidRepository{db: db}
}

func (r *voidRepository) ListLogs(ctx context.Context) ([]VoidLog, error) {
	logs := []VoidLog{}
	query := `
		SELECT l.id, l.member_id, l.member_name, l.member_number, l.action, l.reason,
		       l.admin_id, a.username AS admin_name, l.created_at::text AS created_at,
		       (l.action = 'void'
		        AND l.created_at = (SELECT MAX(x.created_at) FROM vote_void_logs x WHERE x.member_id = l.member_id)
		        AND EXISTS (SELECT 1 FROM votes v WHERE v.member_id = l.member_id AND v.voided_at IS NOT NULL)
		       ) AS can_restore
		FROM vote_void_logs l
		JOIN admins a ON a.id = l.admin_id
		ORDER BY l.created_at DESC
	`
	err := r.db.SelectContext(ctx, &logs, query)
	return logs, err
}

func (r *voidRepository) CountVoided(ctx context.Context) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count, "SELECT COUNT(*) FROM votes WHERE voided_at IS NOT NULL")
	return count, err
}

func (r *voidRepository) SearchVoters(ctx context.Context, search string) ([]Voter, error) {
	voters := []Voter{}
	query := `
		SELECT m.id, m.full_name, m.member_number, k.name AS korda_name
		FROM members m
		JOIN votes v ON v.member_id = m.id AND v.voided_at IS NULL
		LEFT JOIN kordas k ON k.id = m.korda_id
		WHERE m.full_name ILIKE $1 OR m.member_number ILIKE $1
		ORDER BY m.full_name
		LIMIT 10
	`
	err := r.db.SelectContext(ctx, &voters, query, "%"+search+"%")
	return voters, err
}

func tallyActiveVotes(ctx context.Context, q sqlx.QueryerContext) (map[string]int, error) {
	rows, err := q.QueryxContext(ctx, `
		SELECT k.id, COUNT(v.id)
		FROM kandidats k
		LEFT JOIN votes v ON v.kandidat_id = k.id AND v.voided_at IS NULL
		GROUP BY k.id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	counts := map[string]int{}
	for rows.Next() {
		var id string
		var n int
		if err := rows.Scan(&id, &n); err != nil {
			return nil, err
		}
		counts[id] = n
	}
	return counts, rows.Err()
}

func leadersOf(counts map[string]int) string {
	max := 0
	for _, n := range counts {
		if n > max {
			max = n
		}
	}
	if max == 0 {
		return ""
	}
	ids := []string{}
	for id, n := range counts {
		if n == max {
			ids = append(ids, id)
		}
	}
	sort.Strings(ids)
	return strings.Join(ids, ",")
}

func (r *voidRepository) SetVoided(ctx context.Context, memberID, adminID, reason string, void, confirmWinnerChange bool) (bool, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer tx.Rollback()

	var name, number string
	err = tx.QueryRowxContext(ctx, "SELECT full_name, member_number FROM members WHERE id = $1", memberID).Scan(&name, &number)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, ErrVoidMemberNotFound
		}
		return false, err
	}

	before, err := tallyActiveVotes(ctx, tx)
	if err != nil {
		return false, err
	}

	action := "restore"
	updateQuery := "UPDATE votes SET voided_at = NULL WHERE member_id = $1 AND voided_at IS NOT NULL"
	noRowsErr := ErrNoVoidedVote
	if void {
		action = "void"
		updateQuery = "UPDATE votes SET voided_at = now() WHERE member_id = $1 AND voided_at IS NULL"
		noRowsErr = ErrNoVoteToVoid
	}

	res, err := tx.ExecContext(ctx, updateQuery, memberID)
	if err != nil {
		return false, err
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		return false, noRowsErr
	}

	after, err := tallyActiveVotes(ctx, tx)
	if err != nil {
		return false, err
	}

	changed := leadersOf(before) != leadersOf(after)
	if changed && !confirmWinnerChange {
		return true, ErrWinnerWouldChange
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO vote_void_logs (id, member_id, member_name, member_number, action, reason, admin_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		ulid.Make().String(), memberID, name, number, action, reason, adminID,
	)
	if err != nil {
		return false, err
	}

	return changed, tx.Commit()
}

type VoidListResult struct {
	Items       []VoidLog `json:"items"`
	VoidedCount int       `json:"voided_count"`
}

type VoidService interface {
	List(ctx context.Context) (*VoidListResult, error)
	SearchVoters(ctx context.Context, search string) ([]Voter, error)
	Void(ctx context.Context, memberID, adminID, reason string, confirm bool) (bool, error)
	Restore(ctx context.Context, memberID, adminID, reason string, confirm bool) (bool, error)
}

type voidService struct {
	repo VoidRepository
}

func NewVoidService(repo VoidRepository) VoidService {
	return &voidService{repo: repo}
}

func (s *voidService) List(ctx context.Context) (*VoidListResult, error) {
	logs, err := s.repo.ListLogs(ctx)
	if err != nil {
		return nil, err
	}
	count, err := s.repo.CountVoided(ctx)
	if err != nil {
		return nil, err
	}
	return &VoidListResult{Items: logs, VoidedCount: count}, nil
}

func (s *voidService) SearchVoters(ctx context.Context, search string) ([]Voter, error) {
	return s.repo.SearchVoters(ctx, strings.TrimSpace(search))
}

func (s *voidService) Void(ctx context.Context, memberID, adminID, reason string, confirm bool) (bool, error) {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return false, ErrVoidReasonRequired
	}
	return s.repo.SetVoided(ctx, memberID, adminID, reason, true, confirm)
}

func (s *voidService) Restore(ctx context.Context, memberID, adminID, reason string, confirm bool) (bool, error) {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return false, ErrVoidReasonRequired
	}
	return s.repo.SetVoided(ctx, memberID, adminID, reason, false, confirm)
}

type VoidHandler struct {
	service VoidService
}

func NewVoidHandler(service VoidService) *VoidHandler {
	return &VoidHandler{service: service}
}

type voidRequest struct {
	MemberID            string `json:"member_id"`
	Reason              string `json:"reason"`
	ConfirmWinnerChange bool   `json:"confirm_winner_change"`
}

func (h *VoidHandler) List(c echo.Context) error {
	result, err := h.service.List(c.Request().Context())
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "Terjadi kesalahan pada server")
	}
	return response.Success(c, http.StatusOK, "Berhasil mengambil data", result)
}

func (h *VoidHandler) SearchVoters(c echo.Context) error {
	voters, err := h.service.SearchVoters(c.Request().Context(), c.QueryParam("search"))
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "Terjadi kesalahan pada server")
	}
	return response.Success(c, http.StatusOK, "Berhasil mengambil data", voters)
}

func (h *VoidHandler) run(c echo.Context, restore bool) error {
	adminID, _ := c.Get(appMiddleware.ContextKeyAdminID).(string)

	var req voidRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "Format request tidak valid")
	}

	var changed bool
	var err error
	if restore {
		changed, err = h.service.Restore(c.Request().Context(), req.MemberID, adminID, req.Reason, req.ConfirmWinnerChange)
	} else {
		changed, err = h.service.Void(c.Request().Context(), req.MemberID, adminID, req.Reason, req.ConfirmWinnerChange)
	}
	if err != nil {
		return handleVoidError(c, err)
	}

	msg := "Suara berhasil dibatalkan"
	if restore {
		msg = "Suara berhasil dipulihkan"
	}
	return response.Success(c, http.StatusOK, msg, map[string]bool{"winner_changed": changed})
}

func (h *VoidHandler) Void(c echo.Context) error    { return h.run(c, false) }
func (h *VoidHandler) Restore(c echo.Context) error { return h.run(c, true) }

func handleVoidError(c echo.Context, err error) error {
	switch {
	case errors.Is(err, ErrWinnerWouldChange):
		return response.Error(c, http.StatusConflict, err.Error())
	case errors.Is(err, ErrNoVoteToVoid), errors.Is(err, ErrNoVoidedVote), errors.Is(err, ErrVoidReasonRequired):
		return response.Error(c, http.StatusBadRequest, err.Error())
	case errors.Is(err, ErrVoidMemberNotFound):
		return response.Error(c, http.StatusNotFound, err.Error())
	default:
		return response.Error(c, http.StatusInternalServerError, "Terjadi kesalahan pada server")
	}
}

func RegisterVoidRoutes(e *echo.Echo, h *VoidHandler, jwtSecret string) {
	group := e.Group("/api/admin/pemilu-voids",
		appMiddleware.RequireAuth(jwtSecret),
		appMiddleware.RequireRole(auth.RoleAdmin, auth.RoleSuperAdmin),
	)
	group.GET("", h.List)

	group.GET("/voters", h.SearchVoters, appMiddleware.RequireRole(auth.RoleSuperAdmin))
	group.POST("", h.Void, appMiddleware.RequireRole(auth.RoleSuperAdmin))
	group.POST("/restore", h.Restore, appMiddleware.RequireRole(auth.RoleSuperAdmin))
}
