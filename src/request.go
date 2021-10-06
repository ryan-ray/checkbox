package main

import (
	"fmt"
)

// Request represents the incoming request for an images basesd on the
// provided UUID with a requested format and resolution
type Request struct {
	UUID       string `json:"uuid"`
	Format     string `json:"format"`
	Resolution struct {
		Width  int `json:"width"`
		Height int `json:"height"`
	} `json:"resolution"`
}

func (rq Request) Key() string {
	return fmt.Sprintf(
		"%s/%dx%d.%s",
		rq.UUID,
		rq.Resolution.Width,
		rq.Resolution.Height,
		rq.Format,
	)
}

func (rq Request) Path(host, name string) string {
	return fmt.Sprintf(
		"%s/%s/%s",
		host,
		name,
		rq.Key(),
	)
}

func (rq Request) OriginalKey() string {
	return fmt.Sprintf(
		"%s/original.%s",
		rq.UUID,
		rq.Format,
	)
}

func (rq Request) OriginalPath(host, name string) string {
	return fmt.Sprintf(
		"%s/%s/%s",
		host,
		name,
		rq.OriginalKey(),
	)
}
