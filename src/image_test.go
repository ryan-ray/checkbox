package main

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"testing"
)

func TestReformat(t *testing.T) {
	orig := image.NewRGBA(image.Rect(0, 0, 100, 100))

	newReq := func(uuid, format string, width, height int) Request {
		return Request{
			UUID:   uuid,
			Format: format,
			Resolution: struct {
				Width  int `json:"width"`
				Height int `json:"height"`
			}{width, height},
		}
	}

	tests := []struct {
		name string
		req  Request
		orig image.Image
		err  error
	}{
		{"200x200", newReq("200x200", "png", 200, 200), orig, nil},
		{"0x0", newReq("0x0", "png", 0, 0), orig, ErrZeroSizeImage},
		{"1x0", newReq("1x0", "png", 1, 0), orig, ErrZeroSizeImage},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			fmt.Println(tt.req.Resolution)

			err := Reformat(tt.req, tt.orig, &buf)
			if !errors.Is(err, tt.err) {
				t.Errorf("Got %v, want %v", err, tt.err)
			}

			if err != nil {
				return
			}

			if buf.Len() == 0 {
				t.Errorf("buf should not be nil")
			}

			reformatted, ext, err := image.Decode(&buf)
			if err != nil {
				t.Errorf("Got %v, want %v", err, nil)
			}

			if ext != tt.req.Format {
				t.Errorf("Got %s, want %s", ext, tt.req.Format)
			}

			if reformatted.Bounds().Dx() != tt.req.Resolution.Width {
				t.Errorf("Got %d, want %d", reformatted.Bounds().Dx(), tt.req.Resolution.Width)
			}

			if reformatted.Bounds().Dy() != tt.req.Resolution.Height {
				t.Errorf("Got %d, want %d", reformatted.Bounds().Dy(), tt.req.Resolution.Height)
			}
		})
	}
}
