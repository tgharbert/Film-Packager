DROP TABLE IF EXISTS memberships_organizations, doc_comments, documents, memberships, organizations, users;
DROP TYPE IF EXISTS invite_status;

CREATE TABLE "users" (
    "id" UUID PRIMARY KEY,
    "name" VARCHAR(100),
    "email" VARCHAR(100) UNIQUE,
    "password" VARCHAR(255),
    "role" VARCHAR(50)
);

CREATE TABLE "organizations" (
    "id" UUID PRIMARY KEY,
    "name" VARCHAR(50)
);

CREATE TYPE invite_status AS ENUM ('pending', 'accepted', 'rejected', 'revoked');

CREATE TABLE "memberships" (
    "id" UUID PRIMARY KEY,
    "user_id" UUID,
    "organization_id" UUID REFERENCES organizations(id) ON DELETE CASCADE,
    "access_tier" TEXT[] DEFAULT ARRAY['reader'],
    "invite_status" invite_status DEFAULT 'pending'
);

CREATE TABLE "documents" (
    "id" UUID PRIMARY KEY,
    "organization_id" UUID REFERENCES organizations(id) ON DELETE CASCADE,
    "user_id" UUID,
    "file_name" VARCHAR(100),
    "file_type" VARCHAR(50),
    "date" TIMESTAMP,
    "color" VARCHAR(50),
    "status" VARCHAR(50)
);

CREATE TABLE "doc_comments" (
    "id" serial PRIMARY KEY,
    "document_id" UUID,
    "user_id" UUID,
    "comment" VARCHAR(250)
);

CREATE TABLE "memberships_organizations" (
    "membership_id" UUID REFERENCES "memberships" ("id") ON DELETE CASCADE,
    "organization_id" UUID REFERENCES "organizations" ("id") ON DELETE CASCADE,
    PRIMARY KEY ("membership_id", "organization_id")
);

ALTER TABLE "memberships" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id");
ALTER TABLE "documents" ADD FOREIGN KEY ("organization_id") REFERENCES "organizations" ("id");
ALTER TABLE "doc_comments" ADD FOREIGN KEY ("document_id") REFERENCES "documents" ("id");
ALTER TABLE "doc_comments" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id");
ALTER TABLE "documents" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id");
CREATE UNIQUE INDEX unique_org_file_status ON documents (organization_id, file_type) WHERE status IN ('staged', 'locked');

