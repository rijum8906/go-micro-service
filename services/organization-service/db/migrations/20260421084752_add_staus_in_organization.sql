-- Modify "organizations" table
ALTER TABLE "organizations" ADD COLUMN "status" character varying(30) NOT NULL DEFAULT 'active';
