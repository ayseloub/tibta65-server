package pemilu

import (
	"context"
	"errors"
	"time"

	"github.com/oklog/ulid/v2"
)

var ErrValidation = errors.New("data tidak valid")

type KandidatResult struct {
	Kandidat
	Percentage float64 `json:"percentage"`
}

type DashboardResult struct {
	Settings         Settings         `json:"settings"`
	Status           string           `json:"status"`
	TotalMembers     int              `json:"total_members"`
	TotalVotes       int              `json:"total_votes"`
	ParticipationPct float64          `json:"participation_pct"`
	Kandidats        []KandidatResult `json:"kandidats"`
}

type MemberKandidat struct {
	ID         string   `json:"id"`
	FullName   string   `json:"full_name"`
	Visi       string   `json:"visi"`
	Misi       string   `json:"misi"`
	Pangkat    string   `json:"pangkat"`
	VoteCount  *int     `json:"vote_count,omitempty"`
	Percentage *float64 `json:"percentage,omitempty"`
}

type MemberDashboardResult struct {
	Status          string           `json:"status"`
	EndAt           time.Time        `json:"end_at"`
	HasVoted        bool             `json:"has_voted"`
	VotedKandidatID *string          `json:"voted_kandidat_id"`
	Kandidats       []MemberKandidat `json:"kandidats"`
}

type KandidatInput struct {
	FullName string
	Visi     string
	Misi     string
	Pangkat  string
}

type Service interface {
	Dashboard(ctx context.Context) (*DashboardResult, error)
	UpdateSettings(ctx context.Context, startAt, endAt string) (*Settings, error)
	CloseEarly(ctx context.Context) (*Settings, error)
	GetKandidat(ctx context.Context, id string) (*Kandidat, error)
	CreateKandidat(ctx context.Context, in KandidatInput) (*Kandidat, error)
	UpdateKandidat(ctx context.Context, id string, in KandidatInput) (*Kandidat, error)
	DeleteKandidat(ctx context.Context, id string) error
	MemberDashboard(ctx context.Context, memberID string) (*MemberDashboardResult, error)
	CastVote(ctx context.Context, memberID, kandidatID string) error
	ResetPemilu(ctx context.Context) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) ResetPemilu(ctx context.Context) error {
	return s.repo.ResetAll(ctx)
}

func computeStatus(s *Settings) string {
	now := time.Now()
	if s.ClosedEarlyAt != nil {
		return "closed"
	}
	if now.Before(s.StartAt) {
		return "not_started"
	}
	if now.After(s.EndAt) {
		return "closed"
	}
	return "active"
}

func (s *service) Dashboard(ctx context.Context) (*DashboardResult, error) {
	settings, err := s.repo.FindSettings(ctx)
	if err != nil {
		return nil, err
	}

	kandidats, err := s.repo.FindAllKandidat(ctx)
	if err != nil {
		return nil, err
	}

	totalMembers, err := s.repo.CountMembers(ctx)
	if err != nil {
		return nil, err
	}

	totalVotes, err := s.repo.CountVotes(ctx)
	if err != nil {
		return nil, err
	}

	participationPct := 0.0
	if totalMembers > 0 {
		participationPct = float64(totalVotes) / float64(totalMembers) * 100
	}

	results := make([]KandidatResult, len(kandidats))
	for i, k := range kandidats {
		pct := 0.0
		if totalVotes > 0 {
			pct = float64(k.VoteCount) / float64(totalVotes) * 100
		}
		results[i] = KandidatResult{Kandidat: k, Percentage: pct}
	}

	return &DashboardResult{
		Settings:         *settings,
		Status:           computeStatus(settings),
		TotalMembers:     totalMembers,
		TotalVotes:       totalVotes,
		ParticipationPct: participationPct,
		Kandidats:        results,
	}, nil
}

func (s *service) UpdateSettings(ctx context.Context, startAt, endAt string) (*Settings, error) {
	if startAt == "" || endAt == "" {
		return nil, ErrValidation
	}
	return s.repo.UpdateSettings(ctx, startAt, endAt)
}

func (s *service) CloseEarly(ctx context.Context) (*Settings, error) {
	return s.repo.CloseEarly(ctx)
}

func (s *service) GetKandidat(ctx context.Context, id string) (*Kandidat, error) {
	return s.repo.FindKandidatByID(ctx, id)
}

func (s *service) CreateKandidat(ctx context.Context, in KandidatInput) (*Kandidat, error) {
	if in.FullName == "" || in.Visi == "" || in.Misi == "" || in.Pangkat == "" {
		return nil, ErrValidation
	}
	k := &Kandidat{ID: ulid.Make().String(), FullName: in.FullName, Visi: in.Visi, Misi: in.Misi, Pangkat: in.Pangkat}
	if err := s.repo.CreateKandidat(ctx, k); err != nil {
		return nil, err
	}
	return k, nil
}

func (s *service) UpdateKandidat(ctx context.Context, id string, in KandidatInput) (*Kandidat, error) {
	if in.FullName == "" || in.Visi == "" || in.Misi == "" || in.Pangkat == "" {
		return nil, ErrValidation
	}
	k := &Kandidat{ID: id, FullName: in.FullName, Visi: in.Visi, Misi: in.Misi, Pangkat: in.Pangkat}
	if err := s.repo.UpdateKandidat(ctx, k); err != nil {
		return nil, err
	}
	return k, nil
}

func (s *service) DeleteKandidat(ctx context.Context, id string) error {
	return s.repo.DeleteKandidat(ctx, id)
}

func (s *service) MemberDashboard(ctx context.Context, memberID string) (*MemberDashboardResult, error) {
	settings, err := s.repo.FindSettings(ctx)
	if err != nil {
		return nil, err
	}
	status := computeStatus(settings)

	kandidats, err := s.repo.FindAllKandidat(ctx)
	if err != nil {
		return nil, err
	}

	totalVotes := 0
	for _, k := range kandidats {
		totalVotes += k.VoteCount
	}

	votedKandidatID, err := s.repo.FindMemberVote(ctx, memberID)
	if err != nil {
		return nil, err
	}

	showResults := status == "closed"

	result := make([]MemberKandidat, len(kandidats))
	for i, k := range kandidats {
		mk := MemberKandidat{ID: k.ID, FullName: k.FullName, Visi: k.Visi, Misi: k.Misi, Pangkat: k.Pangkat}
		if showResults {
			vc := k.VoteCount
			pct := 0.0
			if totalVotes > 0 {
				pct = float64(k.VoteCount) / float64(totalVotes) * 100
			}
			mk.VoteCount = &vc
			mk.Percentage = &pct
		}
		result[i] = mk
	}

	return &MemberDashboardResult{
		Status: status, EndAt: settings.EndAt, HasVoted: votedKandidatID != nil,
		VotedKandidatID: votedKandidatID, Kandidats: result,
	}, nil
}

func (s *service) CastVote(ctx context.Context, memberID, kandidatID string) error {
	settings, err := s.repo.FindSettings(ctx)
	if err != nil {
		return err
	}
	if computeStatus(settings) != "active" {
		return ErrPemiluNotActive
	}
	return s.repo.CreateVote(ctx, memberID, kandidatID)
}
