package service

import (
	"context"
)

type MMDBRepository interface {
	DownloadCityDb(ctx context.Context, update ...bool) error
	DownloadISPDb(ctx context.Context, update ...bool) error
	CheckCityDbUpdates(ctx context.Context) (bool, error)
	CheckISPDbUpdates(ctx context.Context) (bool, error)
}

type MMDBService struct {
	MMDBRepository
}

func NewMMDBService(rep MMDBRepository) *MMDBService {
	return &MMDBService{
		MMDBRepository: rep,
	}
}

func (s *MMDBService) DownloadCityDb(ctx context.Context, update ...bool) error {
	return s.MMDBRepository.DownloadCityDb(ctx, update...)
}
func (s *MMDBService) DownloadISPDb(ctx context.Context, update ...bool) error {
	return s.MMDBRepository.DownloadISPDb(ctx, update...)
}
func (s *MMDBService) CheckCityDbUpdates(ctx context.Context) (bool, error) {
	return s.MMDBRepository.CheckCityDbUpdates(ctx)
}
func (s *MMDBService) CheckISPDbUpdates(ctx context.Context) (bool, error) {
	return s.MMDBRepository.CheckISPDbUpdates(ctx)
}
