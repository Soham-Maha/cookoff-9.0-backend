CREATE TABLE "user" (
    "id" UUID NOT NULL UNIQUE,
    "submissions" UUID,
    "email" VARCHAR NOT NULL,
    "regNo" VARCHAR NOT NULL,
    "password" VARCHAR NOT NULL,
    "role" CHAR NOT NULL,
    "roundQualified" INTEGER NOT NULL,
    "score" INTEGER,
    "name" VARCHAR NOT NULL,
    PRIMARY KEY("id")
);



CREATE TABLE "questions" (
	"id" UUID NOT NULL UNIQUE,
	"title" VARCHAR,
	"description" TEXT,
	"inputFormat" TEXT,
	"outputFormat" TEXT,
	"points" INTEGER,
	"round" SMALLINT NOT NULL,
	"constraints" TEXT,
	"testcases" UUID,
	PRIMARY KEY("id")
);


CREATE TABLE "submissions" (
	"id" UUID NOT NULL UNIQUE,
	"questionId" UUID NOT NULL,
	"testcases_passed" INTEGER NOT NULL,
	"testcases_failed" INTEGER NOT NULL,
	"runtime" TIME NOT NULL,
	"memory" VARCHAR NOT NULL,
	"sub time" TIMESTAMP NOT NULL,
	"testcases_id" UUID,
	PRIMARY KEY("id")
);


CREATE TABLE "testcases" (
	"id" UUID NOT NULL UNIQUE,
	"expected_output" TEXT,
	"memory" VARCHAR,
	"input" TEXT,
	"hidden" BOOLEAN,
	"runtime" TIME,
	PRIMARY KEY("id")
);


ALTER TABLE "user"
ADD FOREIGN KEY("submissions") REFERENCES "submissions"("id")
ON UPDATE NO ACTION ON DELETE NO ACTION;
ALTER TABLE "submissions"
ADD FOREIGN KEY("questionId") REFERENCES "questions"("id")
ON UPDATE NO ACTION ON DELETE NO ACTION;
ALTER TABLE "questions"
ADD FOREIGN KEY("testcases") REFERENCES "testcases"("id")
ON UPDATE NO ACTION ON DELETE NO ACTION;
ALTER TABLE "submissions"
ADD FOREIGN KEY("testcases_id") REFERENCES "testcases"("id")
ON UPDATE NO ACTION ON DELETE NO ACTION;
