// SPDX-License-Identifier: GPL-2.0-or-later
/*
 * Copyright (C) 2018-2025 SCANOSS.COM
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

package server

import (
	"context"
	"testing"

	common "github.com/scanoss/papi/api/commonv2"
	pb "github.com/scanoss/papi/api/licensesv2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Echo and GetObligations are not implemented. What matters is that calling them
// returns an error rather than panicking: the server used to embed the papi
// LicenseServer interface, whose nil value any unimplemented method dereferenced,
// so one request took the whole process down.
func TestUnimplementedMethodsReturnErrorsInsteadOfPanicking(t *testing.T) {
	server := &LicenseServer{}
	calls := map[string]func() error{
		"Echo": func() error {
			_, err := server.Echo(context.Background(), &common.EchoRequest{})
			return err
		},
		"GetObligations": func() error {
			_, err := server.GetObligations(context.Background(), &pb.LicenseRequest{})
			return err
		},
	}
	for name, call := range calls {
		t.Run(name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("calling %s panicked: %v", name, r)
				}
			}()
			err := call()
			if err == nil {
				t.Fatalf("want an error from %s, which the service does not implement", name)
			}
			if got := status.Code(err); got != codes.Unimplemented {
				t.Errorf("status code = %v, want %v", got, codes.Unimplemented)
			}
		})
	}
}
