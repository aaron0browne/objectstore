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
	"compress/gzip"
	"fmt"
	"io"
	"os"

	"cloud.google.com/go/storage"
)

// Reader is a reader of Object data that meets the io.ReadCloser interface. Use
// Object.NewReader to get a Reader.
type Reader struct {
	io.Reader

	scheme          string
	contentEncoding string
	contentType     string

	f  *os.File
	or *storage.Reader
	gr *gzip.Reader
}

// Close closes the Reader. It must be called when done reading.
func (r *Reader) Close() error {
	if r == nil {
		return nil
	}

	var mErr multiError
	if r.contentEncoding == "gzip" {
		if err := r.gr.Close(); err != nil {
			mErr = append(mErr, fmt.Errorf("%T: %s", r.gr, err))
		}
	}

	switch r.scheme {
	case "gs":
		if err := r.or.Close(); err != nil {
			mErr = append(mErr, fmt.Errorf("%T: %s", r.or, err))
		}
	case "", "file":
		if err := r.f.Close(); err != nil {
			mErr = append(mErr, fmt.Errorf("%T: %s", r.f, err))
		}
	}

	if len(mErr) > 0 {
		return mErr
	}
	return nil
}
