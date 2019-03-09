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

// Writer is a writer of Object data that meets the io.WriteCloser interface. Use
// Object.NewWriter to get a Writer.
type Writer struct {
	io.Writer

	scheme          string
	contentEncoding string
	contentType     string

	f  *os.File
	ow *storage.Writer
	gw *gzip.Writer
}

// Close closes the Writer. It must be called when done writing and the returned
// error should be inspected to determine whether the write was successful.
func (w *Writer) Close() error {
	if w == nil {
		return nil
	}

	var mErr multiError
	if w.contentEncoding == "gzip" {
		if err := w.gw.Close(); err != nil {
			mErr = append(mErr, fmt.Errorf("%T: %s", w.gw, err))
		}
	}

	switch w.scheme {
	case "gs":
		if err := w.ow.Close(); err != nil {
			mErr = append(mErr, fmt.Errorf("%T: %s", w.ow, err))
		}
	case "", "file":
		if err := w.f.Close(); err != nil {
			mErr = append(mErr, fmt.Errorf("%T: %s", w.f, err))
		}
	}

	if len(mErr) > 0 {
		return mErr
	}
	return nil
}
