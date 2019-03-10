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
	"context"
	"fmt"
	"net/url"
	"os"
	"path"
	"strings"

	"cloud.google.com/go/storage"
)

// Object provides methods to create new Readers and Writers of an object. Use
// Store.NewObject to get an Object. ContentType and ContentEncoding must be set
// as desired before calling Object.NewReader or Object.NewWriter, in order to
// read and write properly, based on those attributes.
type Object struct {
	ContentType     string
	ContentEncoding string

	u *url.URL
	o *storage.ObjectHandle
}

func (o *Object) guessContentAttrs() {
	if o == nil || o.u == nil {
		return
	}

	ext := path.Ext(o.u.Path)
	switch {
	case strings.Contains(ext, ".csv"):
		o.ContentType = "text/csv"
	case strings.Contains(ext, ".ndjson"):
		o.ContentType = "application/x-ndjson"
	case strings.Contains(ext, ".json"):
		o.ContentType = "application/json"
	case strings.Contains(ext, ".txt"):
		o.ContentType = "text/plain"
	}

	if strings.Contains(ext, ".gz") || strings.Contains(ext, ".gzip") {
		o.ContentEncoding = "gzip"
	}
}

// URL returns the read-only url.URL of an Object.
func (o *Object) URL() *url.URL {
	if o == nil {
		return nil
	}
	u := &url.URL{}
	*u = *o.u
	return u
}

// Delete removes the object from storage.
func (o *Object) Delete(ctx context.Context) error {
	switch o.u.Scheme {
	case "gs":
		return o.o.Delete(ctx)

	case "", "file":
		return os.Remove(o.u.Path)
	}
	return fmt.Errorf("unsupported url scheme '%s'", o.u.Scheme)
}

// NewReader creates a new Reader of Object's data, handling gzip decoding if
// applicable based on Object.ContentEncoding and storage behavior.
//
// The caller must call Close on the returned Reader.
func (o *Object) NewReader(ctx context.Context) (*Reader, error) {
	if o == nil {
		return nil, nil
	}

	r := &Reader{
		scheme:          o.u.Scheme,
		contentEncoding: o.ContentEncoding,
		contentType:     o.ContentType,
	}

	var err error
	gz := r.contentEncoding == "gzip"
	switch r.scheme {
	case "gs":
		r.or, err = o.o.ReadCompressed(gz).NewReader(ctx)
		if err != nil {
			if r.or != nil {
				r.or.Close()
			}
			return nil, err
		}

		r.Reader = r.or

		// This is a content type applied by GCS to gzipped files
		// uploaded without an explicit one and will force server-side
		// decompression, unless the CacheControl is no-transform.
		if r.contentType == "application/x-gzip" {
			a, err := o.o.Attrs(ctx)
			if err != nil {
				return nil, err
			}
			if a.CacheControl != "no-transform" {
				gz = false
			}
		}

		if gz {
			r.gr, err = gzip.NewReader(r.or)
			if err != nil {
				return nil, err
			}
			r.Reader = r.gr
		}

	case "", "file":
		r.f, err = os.Open(o.u.Path)
		if err != nil {
			if r.f != nil {
				r.f.Close()
			}
			return nil, err
		}

		r.Reader = r.f

		if gz {
			r.gr, err = gzip.NewReader(r.f)
			if err != nil {
				return nil, err
			}
			r.Reader = r.gr
		}

	default:
		return nil, fmt.Errorf("unsupported object uri scheme '%s'", o.u.Scheme)
	}

	return r, nil
}

// NewWriter creates a new Writer of Object's data, handling gzip encoding based
// on Object.ContentEncoding and setting Object.ContentType in storage, if
// applicable. Any existing data in the object will be truncated.
//
// The caller must call Close on the returned Writer.
func (o *Object) NewWriter(ctx context.Context) (*Writer, error) {
	w := &Writer{
		scheme:          o.u.Scheme,
		contentEncoding: o.ContentEncoding,
		contentType:     o.ContentType,
	}

	var err error
	gz := w.contentEncoding == "gzip"
	switch w.scheme {
	case "gs":
		w.ow = o.o.NewWriter(ctx)
		w.ow.ContentType = w.contentType
		w.ow.ContentEncoding = w.contentEncoding
		w.Writer = w.ow

		if gz {
			w.gw = gzip.NewWriter(w.ow)
			w.Writer = w.gw
		}

	case "", "file":
		if o.u.Host != "" && o.u.Host != "localhost" {
			return nil, fmt.Errorf("unsupported file object host '%s'", o.u.Host)
		}

		w.f, err = os.Create(o.u.Path)
		if err != nil {
			if w.f != nil {
				w.f.Close()
			}
			return nil, err
		}

		w.Writer = w.f

		if gz {
			w.gw = gzip.NewWriter(w.f)
			w.Writer = w.gw
		}

	default:
		return nil, fmt.Errorf("unsupported object uri scheme '%s'", o.u.Scheme)
	}

	return w, nil
}
