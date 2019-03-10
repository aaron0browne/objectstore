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

// Package objectstore provides an object store utility that supports local file
// system and Google Cloud Storage objects.
package objectstore

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"cloud.google.com/go/storage"
)

// Option is a function that modifies a Store during initialization.
type Option func(*Store)

// WithGCS is an Option that adds Google Cloud Storage functionality to a Store
// using the passed client.
func WithGCS(cs *storage.Client) Option {
	return func(s *Store) {
		s.cs = cs
	}
}

// Store is used to initialize new Objects. Use New to get a Store.
//
// Stores should be reused instead of created as needed. The methods of Store
// are safe for concurrent use by multiple goroutines.
type Store struct {
	cs *storage.Client
}

// New is the Store initialization function.
func New(opts ...Option) *Store {
	s := &Store{}
	for _, o := range opts {
		o(s)
	}
	return s
}

// NewObject initializes a new Object from the Store. The uri extension is used
// to guess at the correct ContentType and ContentEncoding to set on the Object.
// If the uri scheme is empty, a local file system object is created.
func (s *Store) NewObject(ctx context.Context, uri string) (*Object, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}

	o := &Object{
		u: u,
	}
	o.guessContentAttrs()

	switch u.Scheme {
	case "gs":
		if s.cs == nil {
			return nil, errors.New("WithGCS option required")
		}

		n := u.EscapedPath()
		n = strings.TrimPrefix(n, "/")
		o.o = s.cs.Bucket(u.Hostname()).Object(n)

		a, err := o.o.Attrs(ctx)
		if err == storage.ErrObjectNotExist {
			break
		}
		if err != nil {
			return nil, err
		}
		o.ContentType = a.ContentType
		o.ContentEncoding = a.ContentEncoding

	case "", "file":
		if u.Host != "" && u.Host != "localhost" {
			return nil, fmt.Errorf("unsupported file object host '%s'", u.Host)
		}

	default:
		return nil, fmt.Errorf("unsupported object uri scheme '%s'", u.Scheme)
	}

	return o, nil
}
