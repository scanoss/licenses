package models

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"go.uber.org/zap"
)

type OSADLModelInterface interface {
	GetOSADLByLicenseID(ctx context.Context, s *zap.SugaredLogger, id string) (OSADL, error)
}

type OSADLModel struct {
	db *sqlx.DB
}

// YesNoBool scans the strings "Yes"/"No" (and bool/[]byte forms) from the database into a bool.
type YesNoBool bool

func (b *YesNoBool) Scan(value interface{}) error {
	if value == nil {
		*b = false
		return nil
	}
	switch v := value.(type) {
	case bool:
		*b = YesNoBool(v)
	case string:
		*b = YesNoBool(strings.EqualFold(v, "yes") || strings.EqualFold(v, "true"))
	case []byte:
		*b = YesNoBool(strings.EqualFold(string(v), "yes") || strings.EqualFold(string(v), "true"))
	default:
		return fmt.Errorf("cannot scan %T into YesNoBool", value)
	}
	return nil
}

func (b YesNoBool) Value() (driver.Value, error) {
	if b {
		return "Yes", nil
	}
	return "No", nil
}

// JSONStringSlice is a generic type that handles JSON marshaling/unmarshaling for string slices.
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

type OSADL struct {
	ID                       string         `json:"id" db:"id"`
	Compatibilities          pq.StringArray `json:"compatibilities" db:"compatibilities"`
	Incompatibilities        pq.StringArray `json:"incompatibilities" db:"incompatibilities"`
	DependingCompatibilities pq.StringArray `json:"dependingCompatibilities" db:"depending_compatibilities"`
	CopyleftClause           YesNoBool      `json:"copyleftClause" db:"copyleft_clause"`
	PatentHints              YesNoBool      `json:"patentHints" db:"patent_hints"`
	UseCases                 pq.StringArray `json:"useCases" db:"use_cases"`
}

// NewOSADLModel create a new instance of the OSADL Model.
func NewOSADLModel(db *sqlx.DB) *OSADLModel {
	return &OSADLModel{db: db}
}

// GetOSADLByLicenseID retrieves OSADL data by the given license ID.
func (m *OSADLModel) GetOSADLByLicenseID(ctx context.Context, s *zap.SugaredLogger, licenseID string) (OSADL, error) {
	conn, err := NewConn(ctx, m.db)
	if err != nil {
		return OSADL{}, err
	}
	licenseIDToUpper := strings.ToUpper(licenseID)
	var osadl OSADL
	s.Debugf("LicenseDetail ID: %v", licenseIDToUpper)
	err = conn.QueryRowxContext(ctx,
		"SELECT * FROM osadl WHERE UPPER(id) = $1", licenseIDToUpper).StructScan(&osadl)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		s.Errorf("Error: Failed to query 'osadl' table for %v: %#v", licenseIDToUpper, err)
		return OSADL{}, fmt.Errorf("failed to query the 'osadl' table: %v", err)
	}
	return osadl, nil
}
