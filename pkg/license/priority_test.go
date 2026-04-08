// SPDX-License-Identifier: GPL-2.0-or-later
/*
 * Copyright (C) 2026 SCANOSS.COM
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 2 of the License, or
 * (at your option) any later version.
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

package license

import (
	"testing"

	models "scanoss.com/licenses/pkg/model"
)

func TestPickLicensesByPriority(t *testing.T) {
	tests := []struct {
		name           string
		licenses       []models.PurlLicense
		sourcePriority []int16
		want           []models.PurlLicense
	}{
		{
			name:           "nil licenses returns nil",
			licenses:       nil,
			sourcePriority: []int16{1, 2},
			want:           nil,
		},
		{
			name:           "nil priority returns nil",
			licenses:       []models.PurlLicense{{SourceID: 1, LicenseID: 100}},
			sourcePriority: nil,
			want:           nil,
		},
		{
			name:           "single source single row",
			licenses:       []models.PurlLicense{{SourceID: 1, LicenseID: 100}},
			sourcePriority: []int16{1},
			want:           []models.PurlLicense{{SourceID: 1, LicenseID: 100}},
		},
		{
			name: "highest priority source wins over lower",
			licenses: []models.PurlLicense{
				{SourceID: 2, LicenseID: 200},
				{SourceID: 1, LicenseID: 100},
			},
			sourcePriority: []int16{1, 2},
			want:           []models.PurlLicense{{SourceID: 1, LicenseID: 100}},
		},
		{
			name: "falls through to next priority on miss",
			licenses: []models.PurlLicense{
				{SourceID: 3, LicenseID: 300},
			},
			sourcePriority: []int16{1, 2, 3},
			want:           []models.PurlLicense{{SourceID: 3, LicenseID: 300}},
		},
		{
			name: "ignores sources not in priority list",
			licenses: []models.PurlLicense{
				{SourceID: 99, LicenseID: 999},
				{SourceID: 1, LicenseID: 100},
			},
			sourcePriority: []int16{1, 2},
			want:           []models.PurlLicense{{SourceID: 1, LicenseID: 100}},
		},
		{
			name: "returns all rows for the winning source",
			licenses: []models.PurlLicense{
				{SourceID: 1, LicenseID: 100},
				{SourceID: 1, LicenseID: 101},
				{SourceID: 2, LicenseID: 200},
			},
			sourcePriority: []int16{1, 2},
			want: []models.PurlLicense{
				{SourceID: 1, LicenseID: 100},
				{SourceID: 1, LicenseID: 101},
			},
		},
		{
			name: "no matching source returns nil",
			licenses: []models.PurlLicense{
				{SourceID: 99, LicenseID: 999},
			},
			sourcePriority: []int16{1, 2},
			want:           nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := PickLicensesByPriority(tt.licenses, tt.sourcePriority)
			if len(got) != len(tt.want) {
				t.Fatalf("expected %d rows, got %d (%+v)", len(tt.want), len(got), got)
			}
			for i := range got {
				if got[i].SourceID != tt.want[i].SourceID || got[i].LicenseID != tt.want[i].LicenseID {
					t.Errorf("row %d: expected %+v, got %+v", i, tt.want[i], got[i])
				}
			}
		})
	}
}
