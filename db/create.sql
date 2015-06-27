DROP TABLE IF EXISTS "settings";
CREATE TABLE "settings" (
  "id" bigserial NOT NULL,
  "user" varchar(255) NOT NULL,
  "name" varchar(255) NOT NULL,
  "value" text NOT NULL,
  "created_at" timestamp default NULL,
  CONSTRAINT settings_pkey PRIMARY KEY (id)
) WITH (OIDS=FALSE);

DROP TABLE IF EXISTS "projects";
CREATE TABLE "projects" (
  "id" bigserial NOT NULL,
  "name" varchar(255) NOT NULL,
  "pivotal_id" int NOT NULL,
  "mavenlink_id" int NOT NULL,
  "created_by" varchar(255) NOT NULL,
  "created_at" timestamp default NULL,
  CONSTRAINT projects_pkey PRIMARY KEY (id),
  CONSTRAINT projects_name UNIQUE ("name")
) WITH (OIDS=FALSE);

ALTER TABLE "projects" ADD "mvn_sprint_story_id" varchar(255);
