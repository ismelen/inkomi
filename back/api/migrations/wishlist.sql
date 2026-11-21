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

-- Catalog details of the wished works (cover, synopsis), shared by every user.
CREATE TABLE IF NOT EXISTS wishlist_work (
	workKey TEXT PRIMARY KEY,
	title TEXT NOT NULL DEFAULT '',
	description TEXT NOT NULL DEFAULT '',
	coverUrl TEXT NOT NULL DEFAULT '',
	subjects TEXT NOT NULL DEFAULT '',
	firstPublished TEXT NOT NULL DEFAULT '',
	fetchedAt DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Synopses of the wished works in each language (Wikipedia leads), shared by every user; an empty
-- description records that there was none.
CREATE TABLE IF NOT EXISTS wishlist_synopsis (
	workKey TEXT NOT NULL,
	lang TEXT NOT NULL,
	description TEXT NOT NULL DEFAULT '',
	fetchedAt DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	PRIMARY KEY (workKey, lang)
);
