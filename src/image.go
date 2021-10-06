package main

import (
	"errors"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"

	"github.com/nfnt/resize"
)

var ErrZeroSizeImage = errors.New("cannot have an image resolution of 0")

// Reformat creates a copy of the original image and reformats it according to
// the provided Request. The result is then written to the io.Writer.
//
// Returns a wrapped error resulting from calls to png/jpeg/gif Encode.
func Reformat(req Request, orig image.Image, w io.Writer) error {

	if req.Resolution.Width == 0 || req.Resolution.Height == 0 {
		return fmt.Errorf(
			"requested resolution %dx%d ; %w",
			req.Resolution.Width,
			req.Resolution.Height,
			ErrZeroSizeImage,
		)
	}

	sized := resize.Resize(
		uint(req.Resolution.Width),
		uint(req.Resolution.Height),
		orig,
		resize.MitchellNetravali,
	)

	var err error
	switch req.Format {
	case "png":
		err = png.Encode(w, sized)
	case "jpeg", "jpg":
		err = jpeg.Encode(w, sized, nil)
	case "gif":
		err = gif.Encode(w, sized, nil)
	}

	if err != nil {
		return fmt.Errorf("could not reformat image ; %w", err)
	}

	return nil
}
