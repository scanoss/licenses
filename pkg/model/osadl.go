package models

import (
	"context"
	"database/sql"
	"database/sql/driver"
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
//
// The mining `osadl` table kept the three compatibility lists as PostgreSQL text
// arrays on the row. Serving normalises each into its own table - OSADL_compat,
// OSADL_incompat and OSADL_depending - so they are read as three extra queries
// and reassembled here. The table names are mixed-case and must stay quoted.
func (m *OSADLModel) GetOSADLByLicenseID(ctx context.Context, s *zap.SugaredLogger, licenseID string) (OSADL, error) {
	licenseIDToUpper := strings.ToUpper(licenseID)
	s.Debugf("LicenseDetail ID: %v", licenseIDToUpper)

	var row struct {
		ID          string         `db:"id"`
		Copyleft    sql.NullString `db:"copyleft"`
		PatentHints sql.NullBool   `db:"patent_hints"`
	}
	err := m.db.GetContext(ctx, &row,
		m.db.Rebind(`SELECT id, copyleft, patent_hints FROM "OSADL" WHERE UPPER(id) = ?`), licenseIDToUpper)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return OSADL{}, nil
	case err != nil:
		s.Errorf("Error: Failed to query 'OSADL' table for %v: %#v", licenseIDToUpper, err)
		return OSADL{}, fmt.Errorf("failed to query the 'OSADL' table: %w", err)
	}

	osadl := OSADL{
		ID:             row.ID,
		CopyleftClause: YesNoBool(strings.EqualFold(row.Copyleft.String, "yes") || strings.EqualFold(row.Copyleft.String, "true")),
		PatentHints:    YesNoBool(row.PatentHints.Bool),
	}
	lists := []struct {
		table, column string
		dest          *pq.StringArray
	}{
		{"OSADL_compat", "compat_id", &osadl.Compatibilities},
		{"OSADL_incompat", "incompat_id", &osadl.Incompatibilities},
		{"OSADL_depending", "depend_id", &osadl.DependingCompatibilities},
	}
	for _, l := range lists {
		var values []string
		q := m.db.Rebind(fmt.Sprintf(`SELECT %s FROM "%s" WHERE UPPER(id) = ? ORDER BY %s`, l.column, l.table, l.column))
		if err = m.db.SelectContext(ctx, &values, q, licenseIDToUpper); err != nil {
			s.Errorf("Error: Failed to query '%s' for %v: %#v", l.table, licenseIDToUpper, err)
			return OSADL{}, fmt.Errorf("failed to query the '%s' table: %w", l.table, err)
		}
		*l.dest = values
	}
	return osadl, nil
}
