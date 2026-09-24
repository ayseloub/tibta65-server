package gpaboutsettings

import "context"

type UpdateInput struct {
	Title       string
	Description string
	Stat1       string
	Stat2       string
}

type Service interface {
	Get(ctx context.Context) (*GPAboutSettings, error)
	Update(ctx context.Context, in UpdateInput) (*GPAboutSettings, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) Get(ctx context.Context) (*GPAboutSettings, error) {
	return s.repo.Get(ctx)
}

func (s *service) Update(ctx context.Context, in UpdateInput) (*GPAboutSettings, error) {
	settings := &GPAboutSettings{Title: in.Title, Description: in.Description, Stat1: in.Stat1, Stat2: in.Stat2}
	if err := s.repo.Update(ctx, settings); err != nil {
		return nil, err
	}
	return settings, nil
}
