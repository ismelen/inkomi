-- Wishlists of books and authors (wishlist.db).

CREATE TABLE IF NOT EXISTS wishlist_item (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	userId INTEGER NOT NULL,
	kind TEXT NOT NULL CHECK (kind IN ('book', 'author')),
	title TEXT NOT NULL,
	author TEXT NOT NULL DEFAULT '',
	workKey TEXT NOT NULL DEFAULT '',
	authorKey TEXT NOT NULL DEFAULT '',
	-- what makes two items the same: the Open Library key, or the normalized title and author
	dedupKey TEXT NOT NULL,
	createdAt DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	UNIQUE (userId, kind, dedupKey)
);
CREATE INDEX IF NOT EXISTS wishlist_item_created ON wishlist_item (createdAt);
