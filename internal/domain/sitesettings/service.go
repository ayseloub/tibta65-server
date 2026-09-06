package sitesettings

import "context"

type UpdateInput struct {
	Email        string
	Phone        string
	YoutubeURL   string
	InstagramURL string
	FacebookURL  string
	Address      string
	MapsEmbedURL string
	MapsLink     string
}

type Service interface {
	Get(ctx context.Context) (*Settings, error)
	Update(ctx context.Context, in UpdateInput) (*Settings, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func strPtr(v string) *string {
	if v == "" {
		return nil
	}
	return &v
}

func (s *service) Get(ctx context.Context) (*Settings, error) {
	return s.repo.Find(ctx)
}

func (s *service) Update(ctx context.Context, in UpdateInput) (*Settings, error) {
	settings := &Settings{
		Email: strPtr(in.Email), Phone: strPtr(in.Phone), YoutubeURL: strPtr(in.YoutubeURL), InstagramURL: strPtr(in.InstagramURL),
		FacebookURL: strPtr(in.FacebookURL), Address: strPtr(in.Address),
		MapsEmbedURL: strPtr(in.MapsEmbedURL), MapsLink: strPtr(in.MapsLink),
	}
	if err := s.repo.Update(ctx, settings); err != nil {
		return nil, err
	}
	return settings, nil
}
