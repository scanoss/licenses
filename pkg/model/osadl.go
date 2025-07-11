package models

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
	"strings"
)

type OSADLModelInterface interface {
	GetOSADLByLicenseId(ctx context.Context, s *zap.SugaredLogger, id string) (OSADL, error)
}

type OSADLModel struct {
	db *sqlx.DB
}

// JSONStringSlice is a generic type that handles JSON marshaling/unmarshaling for string slices
type JSONStringSlice []string

func (s *JSONStringSlice) Scan(value interface{}) error {
	if value == nil {
		*s = nil
		return nil
	}

	str, ok := value.(string)
	if !ok {
		*s = nil
		return nil
	}

	if str == "" {
		*s = nil
		return nil
	}

	var result []string
	if err := json.Unmarshal([]byte(str), &result); err != nil {
		*s = nil
		return nil
	}

	*s = JSONStringSlice(result)
	return nil
}

func (s JSONStringSlice) Value() (driver.Value, error) {
	if len(s) == 0 {
		return nil, nil
	}

	data, err := json.Marshal([]string(s))
	if err != nil {
		return nil, err
	}

	return string(data), nil
}

// Now define your specific types as aliases
type Compatibilities JSONStringSlice
type Incompatibilities JSONStringSlice
type DependingCompatibilities JSONStringSlice
type UseCases JSONStringSlice

type OSADL struct {
	ID                       int32           `json:"id" db:"id"`
	LicenseId                string          `json:"licenseId" db:"license_id"`
	Compatibilities          JSONStringSlice `json:"compatibilities" db:"compatibilities"`
	Incompatibilities        JSONStringSlice `json:"incompatibilities" db:"incompatibilities"`
	DependingCompatibilities JSONStringSlice `json:"dependingCompatibilities" db:"depending_compatibilities"`
	CopyleftClause           bool            `json:"copyleftClause" db:"copyleft_clause"`
	PatentHints              bool            `json:"patentHints" db:"patent_hints"`
	UseCases                 JSONStringSlice `json:"useCases" db:"use_cases"`
}

// NewOSADLModel create a new instance of the OSADL Model.
func NewOSADLModel(db *sqlx.DB) *OSADLModel {
	return &OSADLModel{db: db}
}

// GetLicenseByID retrieves license data by the given row ID.
func (m *OSADLModel) GetOSADLByLicenseId(ctx context.Context, s *zap.SugaredLogger, licenseId string) (OSADL, error) {
	conn, err := NewConn(ctx, m.db)
	if err != nil {
		return OSADL{}, err
	}
	licenseIDToUpper := strings.ToUpper(licenseId)
	var osadl OSADL
	s.Debugf("License ID: %v", licenseIDToUpper)
	err = conn.QueryRowxContext(ctx,
		"SELECT * FROM osadl WHERE UPPER(license_id) = $1", licenseIDToUpper).StructScan(&osadl)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		s.Errorf("Error: Failed to query 'osadl' table for %v: %#v", licenseIDToUpper, err)
		return OSADL{}, fmt.Errorf("failed to query the 'osadl' table: %v", err)
	}
	return osadl, nil
}
