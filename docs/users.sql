CREATE TABLE IF NOT EXISTS "users" (
	"id" INTEGER NOT NULL,
	"email" TEXT,
	"passord_hash" TEXT,
	"isGuest" BOOLEAN NOT NULL DEFAULT false,
	"notifyToken" TEXT,
	"cloudFolder" TEXT,
	"cloudTokenEnc" TEXT,
	PRIMARY KEY("id")
);

CREATE TABLE IF NOT EXISTS "roles" (
	"id" INTEGER NOT NULL,
	"name" TEXT NOT NULL UNIQUE,
	PRIMARY KEY("id")
);

CREATE TABLE IF NOT EXISTS "user_roles" (
	"userId" INTEGER NOT NULL,
	"roleId" INTEGER NOT NULL,
	PRIMARY KEY("userId", "roleId"),
	FOREIGN KEY ("userId") REFERENCES "users"("id")
	ON UPDATE NO ACTION ON DELETE NO ACTION,
	FOREIGN KEY ("roleId") REFERENCES "roles"("id")
	ON UPDATE NO ACTION ON DELETE NO ACTION,
	CONSTRAINT "user_roles_unique_0" UNIQUE ("userId", "roleId")
);

CREATE TABLE IF NOT EXISTS "sessions" (
	"id" INTEGER NOT NULL,
	"userId" INTEGER NOT NULL,
	"refresh_hash" TEXT NOT NULL,
	"deviceId" TEXT NOT NULL UNIQUE,
	PRIMARY KEY("id"),
	FOREIGN KEY ("userId") REFERENCES "users"("id")
	ON UPDATE NO ACTION ON DELETE NO ACTION
);
