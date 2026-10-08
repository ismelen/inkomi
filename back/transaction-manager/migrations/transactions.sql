CREATE TABLE IF NOT EXISTS sources (
	id TEXT NOT NULL,
	userId INTEGER NOT NULL,

	size INTEGER,
	filename TEXT NOT NULL,
	title TEXT NOT NULL,
	kepubify BOOLEAN NOT NULL FALSE,

	type TEXT NOT NULL CHECK (type IN (
		'file',
		'folder',
		'library'
	))

	status TEXT NOT NULL DEFAULT 'pending_upload'
		CHECK (status IN (
			'pending_upload',
			'queued',
			'processing',
			'waiting_join',
			'joining',
			'kepubifying',
			'sent',
			'failed',
			'done'
		)),

	error TEXT,

	createdAt DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updatedAt DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	completedAt DATETIME,

	shouldJoin BOOLEAN NOT NULL DEFAULT FALSE,
	readingDirection TEXT NOT NULL DEFAULT 'ltr'
    CHECK (readingDirection IN ('ltr', 'rtl'))

	folderId TEXT,

	configHash TEXT,

	PRIMARY KEY (id, userId),

	FOREIGN KEY (folderId, userId)
		REFERENCES sources(id, userId)
		ON UPDATE CASCADE
		ON DELETE CASCADE,

	FOREIGN KEY (configHash)
		REFERENCES configs(hash)
		ON UPDATE NO ACTION
		ON DELETE NO ACTION
);

CREATE TABLE IF NOT EXISTS configs (
	hash TEXT PRIMARY KEY,
	data TEXT NOT NULL,
	createdAt DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	lastUsed DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
