package models

import (
	"strings"
	"time"

	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/shared/strutil"
)

const maxTextLength = 300

type ItemKind string

const (
	ItemKindBook   ItemKind = "book"
	ItemKindAuthor ItemKind = "author"
)

func (k ItemKind) IsValid() bool {
	return k == ItemKindBook || k == ItemKindAuthor
}

// Item is a book or an author a user wants. Keys are Open Library keys (OL…W for works, OL…A
// for authors) when the client picked a suggestion; they are empty for free text.
type Item struct {
	Id        int64
	UserId    int
	Kind      ItemKind
	Title     string
	Author    string
	WorkKey   string
	AuthorKey string
	CreatedAt time.Time
}

// NewItem validates and trims a wished item. A book needs a title; an author needs its name,
// which goes in Title.
func NewItem(userId int, kind ItemKind, title, author, workKey, authorKey string) (*Item, error) {
	if !kind.IsValid() {
		return nil, ErrInvalidKind
	}
	item := &Item{
		UserId:    userId,
		Kind:      kind,
		Title:     strings.TrimSpace(title),
		Author:    strings.TrimSpace(author),
		WorkKey:   strings.TrimSpace(workKey),
		AuthorKey: strings.TrimSpace(authorKey),
	}
	if item.Title == "" {
		return nil, ErrEmptyTitle
	}
	if len(item.Title) > maxTextLength || len(item.Author) > maxTextLength {
		return nil, ErrTooLong
	}
	if kind == ItemKindAuthor {
		item.Author, item.WorkKey = "", ""
	}
	return item, nil
}

// DedupKey identifies the same wish: its Open Library key or its normalized title and author.
func (i *Item) DedupKey() string {
	switch {
	case i.Kind == ItemKindBook && i.WorkKey != "":
		return "ol:" + i.WorkKey
	case i.Kind == ItemKindAuthor && i.AuthorKey != "":
		return "ol:" + i.AuthorKey
	}
	key := strutil.Normalize(i.Title)
	if i.Author != "" {
		key += "|" + strutil.Normalize(i.Author)
	}
	return key
}
