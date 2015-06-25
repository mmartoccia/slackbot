CREATE TABLE "settings" (
  "id" bigserial NOT NULL,
  "user" varchar(255) NOT NULL,
  "name" varchar(255) NOT NULL,
  "value" text NOT NULL,
  "created_at" timestamp default NULL,
  CONSTRAINT settings_pkey PRIMARY KEY (id)
) WITH (OIDS=FALSE);
