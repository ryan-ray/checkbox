package main

import (
	"errors"
	"fmt"
	"image"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
)

var (
	ErrNoSuchKey    = errors.New(s3.ErrCodeNoSuchKey)
	ErrNoSuchBucket = errors.New(s3.ErrCodeNoSuchBucket)

	allowedImageFormats = []string{"png", "jpeg", "jpg", "gif"}
)

// S3Client is an interface that we can use to enable these functions to be
// testable. The AWS SDK doens't provide this interface itself.
type S3Client interface {
	GetObject(*s3.GetObjectInput) (*s3.GetObjectOutput, error)
}

// GetObject is a wrapper function around the s3.S3.GetObject function
// that returns sentinal errors that are easier to deal with using native error
// handling functions.
func GetObject(c S3Client, bucket, key string) (*s3.GetObjectOutput, error) {
	obj, err := c.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})

	var aerr awserr.Error
	if errors.As(err, &aerr) {
		switch aerr.Code() {
		case s3.ErrCodeNoSuchKey:
			return nil, ErrNoSuchKey
		case s3.ErrCodeNoSuchBucket:
			return nil, ErrNoSuchBucket
		default:
			return nil, fmt.Errorf("other error ; %w", aerr)
		}
	}

	return obj, nil
}

// GetOriginalImage looks for our original image. If the original image doesn't
// exist in the format in our Request, look for it in any format available
// and convert to our requested format.
//
// Returns an image.Image and and error, all of which are considered fatal
func GetOriginalImage(c S3Client, r Request) (image.Image, error) {
	original, err := GetObject(c, bucketName, r.OriginalKey())
	// Haven't found original so we see if we have it in another format
	if err != nil {
		for _, format := range allowedImageFormats {
			if format == r.Format {
				// We already know this doesn't exist. Skip
				continue
			}

			key := fmt.Sprintf("%s/original.%s", r.UUID, format)
			obj, err := GetObject(c, bucketName, key)
			if err != nil {
				// Any non nil error means the file doesn't exist. Skip
				continue
			}

			// If we are here, we have a hit, break
			original = obj
			break
		}
	}

	if original == nil {
		path := r.OriginalPath(bucketHost, bucketName)
		return nil, fmt.Errorf("could not get original file %s ; %w", path, ErrNoSuchKey)
	}
	defer original.Body.Close()

	img, _, err := image.Decode(original.Body)
	if err != nil {
		return nil, fmt.Errorf("could not decode original image ; %w", err)
	}

	return img, nil
}
