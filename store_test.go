// Copyright (C) 2019 Aaron Browne
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// A copy of the license can be found in the LICENSE file and at
//         https://www.gnu.org/licenses/agpl.html

package objectstore

import (
	"context"
	"testing"
)

func TestGCSObjectNoClientError(t *testing.T) {
	s := New()
	_, err := s.NewObject(context.Background(), "gs://foo/bar")
	if err == nil {
		t.Fatal("new GCS object initialization without client should error")
	}
}
