package models

import "time"

type Section string

const (
	// SectionForYou is what is closest to the user's taste.
	SectionForYou Section = "for_you"
	// SectionNew is what was published recently, closest to the user's taste.
	SectionNew Section = "new"
	// SectionDiscover is close to the user's taste but by authors they don't know.
	SectionDiscover Section = "discover"
)

func (s Section) IsValid() bool {
	return s == SectionForYou || s == SectionNew || s == SectionDiscover
}

func Sections() []Section {
	return []Section{SectionForYou, SectionNew, SectionDiscover}
}

// Recommendation is a work suggested to a user in a section; Reason is the user's work that is
// closest to it ("because you sent X").
type Recommendation struct {
	Section    Section
	Rank       int
	Work       Work
	Score      float64
	Reason     *Work
	ComputedAt time.Time
}
