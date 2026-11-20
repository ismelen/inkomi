package models

import "time"

// Release is the latest published version of the Android app.
type Release struct {
	Version     Version
	Notes       string
	PublishedAt time.Time
	APK         ReleaseAsset
}

type ReleaseAsset struct {
	Name string
	URL  string
	Size int64
	// SHA256 is empty when the source doesn't publish it.
	SHA256 string
}
