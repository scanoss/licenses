package models

import (
	"context"
	"fmt"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type ComponentLicenseModelInterface interface {
	GetComponentLicenses(ctx context.Context, s *zap.SugaredLogger, componentMD5 string) ([]ComponentLicense, error)
}

type ComponentLicenseModel struct {
	db *sqlx.DB
}

type ComponentLicense struct {
	ID              int32  `json:"id" db:"id"`
	PurlMD5         string `json:"purlMD5" db:"purl_md5"`
	Source          int32  `json:"source" db:"source"`
	Statement       string `json:"statement" db:"statement"`
	SPDXIdentifiers string `json:"SPDXIdentifiers" db:"spdx_id"`
}

// NewComponentLicenseModel create a new instance of the License Model.
func NewComponentLicenseModel(db *sqlx.DB) *ComponentLicenseModel {
	return &ComponentLicenseModel{db: db}
}

// GetComponentLicenses retrieves license data by the given row ID.
func (m *ComponentLicenseModel) GetComponentLicenses(ctx context.Context, s *zap.SugaredLogger, componentMD5 string) ([]ComponentLicense, error) {
	conn, err := NewConn(ctx, m.db)
	if err != nil {
		return []ComponentLicense{}, err
	}
	var componentLicenses []ComponentLicense
	rows, err := conn.QueryxContext(ctx,
		"SELECT id, purl_md5, source, statement, spdx_id FROM component_licenses WHERE purl_md5 = $1",
		componentMD5)
	if err != nil {
		s.Errorf("Error: Failed to query component_licenses table for %v: %#v", componentMD5, err)
		return []ComponentLicense{}, fmt.Errorf("failed to query the component_licenses table: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var license ComponentLicense
		err := rows.StructScan(&license)
		if err != nil {
			s.Errorf("Error: Failed to scan component_license row: %#v", err)
			return []ComponentLicense{}, fmt.Errorf("failed to scan component_license row: %v", err)
		}
		componentLicenses = append(componentLicenses, license)
	}

	if err = rows.Err(); err != nil {
		s.Errorf("Error: Row iteration error: %#v", err)
		return []ComponentLicense{}, fmt.Errorf("row iteration error: %v", err)
	}

	return componentLicenses, nil
}
