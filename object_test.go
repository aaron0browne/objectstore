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
	"net/url"
	"testing"
)

func TestGuessContentAttrs(t *testing.T) {
	testCases := []struct{
		name    string
		uri     string
		expType string
		expEnc  string
	}{
		{
			name: "NoExtNoAttrs",
			uri:  "foobar",
		},
		{
			name: "GzipEncoding",
			uri:  "foobar.gzip",
			expEnc: "gzip",
		},
		{
			name: "GzEncoding",
			uri:  "foobar.gz",
			expEnc: "gzip",
		},
		{
			name: "CsvType",
			uri:  "foobar.csv",
			expType: "text/csv",
		},
		{
			name: "NdjsonType",
			uri:  "foobar.ndjson",
			expType: "application/x-ndjson",
		},
		{
			name: "JsonType",
			uri:  "foobar.json",
			expType: "application/json",
		},
		{
			name: "TxtType",
			uri:  "foobar.txt",
			expType: "text/plain",
		},
		{
			name: "GzipNdjsonType",
			uri:  "foobar.ndjson.gzip",
			expType: "application/x-ndjson",
			expEnc: "gzip",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			u, err := url.Parse(tc.uri)
			if err != nil {
				t.Fatal(err)
			}
			o := Object{u: u}
			o.guessContentAttrs()
			if o.ContentType != tc.expType {
				t.Errorf("expected ContentType not guessed: %s != %s", tc.expType, o.ContentType)
			}
			if o.ContentEncoding != tc.expEnc {
				t.Errorf("expected ContentEncoding not guessed: %s != %s", tc.expEnc, o.ContentEncoding)
			}
		})
	}
}
