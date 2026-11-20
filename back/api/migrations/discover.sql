-- Wishlist and book recommendations (discover.db).

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

-- Works known to the recommender: the users' books (origin user) and the candidates found
-- through their authors and subjects (origin candidate).
CREATE TABLE IF NOT EXISTS catalog_work (
	workKey TEXT PRIMARY KEY,
	title TEXT NOT NULL,
	authors TEXT NOT NULL DEFAULT '[]',
	authorKeys TEXT NOT NULL DEFAULT '[]',
	subjects TEXT NOT NULL DEFAULT '[]',
	languages TEXT NOT NULL DEFAULT '[]',
	firstPublishYear INTEGER,
	coverId INTEGER,
	description TEXT NOT NULL DEFAULT '',
	origin TEXT NOT NULL CHECK (origin IN ('user', 'candidate')),
	embedding BLOB,
	embeddedAt DATETIME,
	createdAt DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	lastSeenAt DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS catalog_work_unembedded ON catalog_work (embeddedAt) WHERE embedding IS NULL;

-- What each user sent or wished (the signals of the recommender), matched against Open Library:
-- matchedKey is the work (kind book) or the author (kind author) once status is matched.
CREATE TABLE IF NOT EXISTS user_signal (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	userId INTEGER NOT NULL,
	ref TEXT NOT NULL,
	source TEXT NOT NULL CHECK (source IN ('sent', 'wishlist')),
	kind TEXT NOT NULL CHECK (kind IN ('book', 'author')),
	title TEXT NOT NULL,
	author TEXT NOT NULL DEFAULT '',
	status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'matched', 'unmatched')),
	attempts INTEGER NOT NULL DEFAULT 0,
	matchedKey TEXT NOT NULL DEFAULT '',
	createdAt DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	UNIQUE (userId, ref)
);
CREATE INDEX IF NOT EXISTS user_signal_status ON user_signal (status, attempts);

CREATE TABLE IF NOT EXISTS recommendation (
	userId INTEGER NOT NULL,
	section TEXT NOT NULL CHECK (section IN ('for_you', 'new', 'discover')),
	rank INTEGER NOT NULL,
	workKey TEXT NOT NULL,
	score REAL NOT NULL,
	reasonWorkKey TEXT NOT NULL DEFAULT '',
	computedAt DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	PRIMARY KEY (userId, section, rank)
);

CREATE TABLE IF NOT EXISTS dismissed_work (
	userId INTEGER NOT NULL,
	workKey TEXT NOT NULL,
	createdAt DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	PRIMARY KEY (userId, workKey)
);

-- Users whose recommendations must be computed again.
CREATE TABLE IF NOT EXISTS stale_user (
	userId INTEGER PRIMARY KEY,
	markedAt DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS sync_state (
	name TEXT PRIMARY KEY,
	value TEXT NOT NULL
);
