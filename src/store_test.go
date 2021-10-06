package main

import (
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
)

type MockGetObject func(*s3.GetObjectInput) (*s3.GetObjectOutput, error)

type MockS3Client struct {
	getObject MockGetObject
}

func (m MockS3Client) GetObject(in *s3.GetObjectInput) (*s3.GetObjectOutput, error) {
	return m.getObject(in)
}

func TestGetObject(t *testing.T) {
	noKeyGetObject := func(in *s3.GetObjectInput) (*s3.GetObjectOutput, error) {
		return nil, awserr.New(s3.ErrCodeNoSuchKey, "no such key", errors.New("no such key"))
	}

	noBucketGetObject := func(in *s3.GetObjectInput) (*s3.GetObjectOutput, error) {
		return nil, awserr.New(s3.ErrCodeNoSuchBucket, "no such bucket", errors.New("no such bucket"))
	}

	tests := []struct {
		name          string
		getObjectFunc MockGetObject
		err           error
	}{
		{"NoSuchKey", noKeyGetObject, ErrNoSuchKey},
		{"NoSuchBucket", noBucketGetObject, ErrNoSuchBucket},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s3c := MockS3Client{tt.getObjectFunc}
			if _, err := GetObject(s3c, "", ""); !errors.Is(err, tt.err) {
				t.Errorf("Got `%v`, want `%v`", err, tt.err)
			}
		})
	}

}
