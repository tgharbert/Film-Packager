DROP TABLE IF EXISTS users, organizations, memberships, documents, doc_comments, memberships_organizations;

CREATE TABLE "users" (
    "id" serial PRIMARY KEY,
    "name" VARCHAR(100),
    "email" VARCHAR(100) UNIQUE,
    "password" VARCHAR(255),
    "role" VARCHAR(50)
);

CREATE TABLE "organizations" (
    "id" serial PRIMARY KEY,
    "name" VARCHAR(50)
);

CREATE TABLE "memberships" (
    "id" serial PRIMARY KEY,
    "user_id" int,
    "organization_id" int,
    "access_tier" VARCHAR(50)
);

CREATE TABLE "documents" (
    "id" serial PRIMARY KEY,
    "organization_id" int,
    "user_id" int,
    "address" VARCHAR(100),
    "name" VARCHAR(50),
    "date" date,
    "color" VARCHAR(50),
    "status" VARCHAR(50)
);

CREATE TABLE "doc_comments" (
    "id" serial PRIMARY KEY,
    "document_id" int,
    "user_id" int,
    "comment" VARCHAR(250)
);

CREATE TABLE "memberships_organizations" (
    "memberships_organization_id" int,
    "organizations_id" serial,
    PRIMARY KEY ("memberships_organization_id", "organizations_id")
);

ALTER TABLE "memberships" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id");

ALTER TABLE "documents" ADD FOREIGN KEY ("organization_id") REFERENCES "organizations" ("id");

ALTER TABLE "doc_comments" ADD FOREIGN KEY ("document_id") REFERENCES "documents" ("id");

ALTER TABLE "memberships_organizations" ADD FOREIGN KEY ("memberships_organization_id") REFERENCES "memberships" ("organization_id");

ALTER TABLE "memberships_organizations" ADD FOREIGN KEY ("organizations_id") REFERENCES "organizations" ("id");

ALTER TABLE "doc_comments" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id");

ALTER TABLE "documents" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id");

ALTER TABLE "memberships_organizations" ADD FOREIGN KEY ("memberships_organization_id") REFERENCES "memberships" ("id");
