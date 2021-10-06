package main

// Response represents the response containing a URL to the image fetched
// or generated from a provided Request
type Response struct {
	URL string `json:"url"`
}
