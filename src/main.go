package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

// All of these constants are needed for localstack operations, they are
// here as a hacky way to get everything working. In practice these could be
// stored in Parameter Store, or Vault or something similar.
const (
	awsEndpoint = "http://localhost:4566"
	bucketHost  = "http://localhost:4566"
	bucketName  = "checkboximageupload"
)

type ImageRequestHandler func(r Request) (*Response, error)

func Handle(s *session.Session, c *s3.S3) ImageRequestHandler {
	return func(r Request) (*Response, error) {
		res := &Response{}
		path := r.Path(bucketHost, bucketName)

		// Log out our request for debugging purposes
		reqJson, _ := json.MarshalIndent(r, "", "  ")
		log.Printf("REQUEST: %s", reqJson)

		// Check to see if we have a cached version of our image in our
		// requested format and resolution. Return that if we do.
		log.Printf("Looking for %s", path)
		var err error
		if _, err = GetObject(c, bucketName, r.Key()); err == nil {
			log.Printf("Found at %s", r.Path(bucketHost, bucketName))
			res.URL = path

			return res, err
		}

		// If ther error is anything except ErrNoSuchKey, we can't really go any
		// further, so we log and return
		if !errors.Is(err, ErrNoSuchKey) {
			err = fmt.Errorf("error getting cached image ; %w", err)
			log.Printf(err.Error())

			return res, err
		}

		log.Printf("Image %s not cached. Creating...", path)

		// We don't have a cached version of the request, so we need to get the
		// original image and create our desired format
		//
		// If the original doesn't exist we can't do anything, so we log and
		// return
		original, err := GetOriginalImage(c, r)
		if err != nil {
			err = fmt.Errorf("could not get original image ; %w", err)
			log.Printf(err.Error())

			return res, err
		}

		// Reformat our image based on our Request, any errors means we can't
		// go any further so we return
		var buf bytes.Buffer
		if err := Reformat(r, original, &buf); err != nil {
			log.Printf(err.Error())

			return res, err
		}

		// Upload our reformatted image to S3 and return the appropriate Response
		uploader := s3manager.NewUploader(s)
		_, err = uploader.Upload(&s3manager.UploadInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(r.Key()),
			Body:   &buf,
		})
		// If we can't upload then we can't go further. Log and return
		if err != nil {
			err = fmt.Errorf("could not upload to s3 ; %w", err)
			log.Printf(err.Error())

			return res, err
		}

		// If we are here, we should be all good. Log some info and return our
		// successful response
		log.Printf("Image created in bucket `%s` with key `%s`", bucketName, r.Key())

		res.URL = path
		log.Printf("RESPONSE: %s", res)

		return res, nil
	}
}

func main() {
	s := session.Must(session.NewSession(&aws.Config{
		Credentials:      credentials.NewStaticCredentials("test", "test", ""),
		S3ForcePathStyle: aws.Bool(true),
		Region:           aws.String(endpoints.UsEast1RegionID),
		Endpoint:         aws.String(awsEndpoint),
	}))

	c := s3.New(s, &aws.Config{})

	lambda.Start(Handle(s, c))
}
