package usecase

import (
	"context"
	"go.uber.org/zap"
	myconfig "scanoss.com/licenses/pkg/config"
	"scanoss.com/licenses/pkg/interfaces"
)

type LicenseUseCase struct {
	s      *zap.SugaredLogger
	config *myconfig.ServerConfig
}

func NewLicenseUseCase(s *zap.SugaredLogger, config *myconfig.ServerConfig) *LicenseUseCase {
	return &LicenseUseCase{config: config}
}

// GetLicenses
func (d LicenseUseCase) GetLicenses(ctx context.Context, components []interfaces.Component) {

}
