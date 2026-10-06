CREATE TABLE IF NOT EXISTS "sources" (
	"id" INTEGER NOT NULL,
	"userId" VARCHAR NOT NULL,
	"size" NUMERIC,
	"filename" TEXT NOT NULL,
	"title" TEXT NOT NULL,
	"status" TEXT NOT NULL CHECK(status IN ('pending_upload', 'queued', 'processing', 'joining', 'kepubifiing', 'sent', 'failed')),
	"error" TEXT,
	"createdAt" DATETIME NOT NULL,
	"updatedAt" DATETIME NOT NULL,
	"completedAt" DATETIME,
	"join" BOOLEAN NOT NULL DEFAULT false,
	"ltr" BOOLEAN NOT NULL DEFAULT true,
	"parentId" INTEGER NOT NULL,
	"configId" VARCHAR,
	PRIMARY KEY("id", "userId"),
	FOREIGN KEY ("parentId") REFERENCES "sources"("id")
	ON UPDATE CASCADE ON DELETE CASCADE,
	FOREIGN KEY ("configId") REFERENCES "configs"("hash")
	ON UPDATE NO ACTION ON DELETE NO ACTION
);

CREATE TABLE IF NOT EXISTS "configs" (
	"hash" VARCHAR NOT NULL,
	"data" TEXT NOT NULL,
	"createdAt" DATETIME NOT NULL,
	"lastUsed" DATETIME NOT NULL,
	PRIMARY KEY("hash")
);
