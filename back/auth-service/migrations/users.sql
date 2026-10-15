CREATE TABLE IF NOT EXISTS users (
	id INTEGER PRIMARY KEY,
	email TEXT UNIQUE,
	emailVerified BOOLEAN NOT NULL DEFAULT FALSE,
	passwordHash TEXT,
	role TEXT NOT NULL DEFAULT 'user',
	fcmToken TEXT,
	createdAt INTEGER NOT NULL DEFAULT (CAST(strftime('%s', 'now') AS INTEGER))
);

CREATE TABLE IF NOT EXISTS identities (
	provider TEXT NOT NULL,
	subject TEXT NOT NULL,
	userId INTEGER NOT NULL,

	PRIMARY KEY (provider, subject),
	FOREIGN KEY (userId)
		REFERENCES users(id)
		ON UPDATE CASCADE
		ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_identities_user ON identities(userId);

CREATE TABLE IF NOT EXISTS dropbox_links (
	userId INTEGER PRIMARY KEY,
	accountId TEXT NOT NULL,
	email TEXT NOT NULL DEFAULT '',
	refreshTokenEnc TEXT NOT NULL,
	accessTokenEnc TEXT,
	accessExpiresAt INTEGER,
	folder TEXT NOT NULL DEFAULT '',

	FOREIGN KEY (userId)
		REFERENCES users(id)
		ON UPDATE CASCADE
		ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS sessions (
	id INTEGER PRIMARY KEY,
	userId INTEGER NOT NULL,
	deviceId TEXT NOT NULL,
	refreshHash TEXT NOT NULL UNIQUE,
	expiresAt INTEGER NOT NULL,

	UNIQUE (userId, deviceId),
	FOREIGN KEY (userId)
		REFERENCES users(id)
		ON UPDATE CASCADE
		ON DELETE CASCADE
);
