// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib" // required for SQL access
	"github.com/jmoiron/sqlx"
	migrate "github.com/rubenv/sql-migrate"
)

// Config defines the options that are used when connecting to a PostgreSQL instance
type Config struct {
	Host        string
	Port        string
	User        string
	Pass        string
	Name        string
	SSLMode     string
	SSLCert     string
	SSLKey      string
	SSLRootCert string
}

// Connect creates a connection to the PostgreSQL instance and applies any
// unapplied database migrations. A non-nil error is returned to indicate failure.
func Connect(cfg Config) (*sqlx.DB, error) {
	url := fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=%s sslcert=%s sslkey=%s sslrootcert=%s", cfg.Host, cfg.Port, cfg.User, cfg.Name, cfg.Pass, cfg.SSLMode, cfg.SSLCert, cfg.SSLKey, cfg.SSLRootCert)

	db, err := sqlx.Open("pgx", url)
	if err != nil {
		return nil, err
	}

	if err := migrateDB(db); err != nil {
		return nil, err
	}
	return db, nil
}

func migrateDB(db *sqlx.DB) error {
	migrations := &migrate.MemoryMigrationSource{
		Migrations: []*migrate.Migration{
			{
				Id: "auth_1",
				Up: []string{
					`CREATE TABLE IF NOT EXISTS keys (
						id          VARCHAR(254) NOT NULL,
						type        SMALLINT,
						subject     VARCHAR(254) NOT NULL,
						issuer_id   UUID NOT NULL,
						issued_at   TIMESTAMP NOT NULL,
						expires_at  TIMESTAMP,
						PRIMARY KEY (id, issuer_id)
					)`,
				},
				Down: []string{
					`DROP TABLE IF EXISTS keys`,
				},
			},
			{
				Id: "auth_2",
				Up: []string{
					`CREATE TABLE IF NOT EXISTS orgs (
							id          UUID UNIQUE NOT NULL,
							owner_id    UUID,
							name        VARCHAR(254) NOT NULL,
							description VARCHAR(1024),
							metadata    JSONB,
							created_at  TIMESTAMPTZ,
							updated_at  TIMESTAMPTZ,
							PRIMARY KEY (id, owner_id)
						 )`,
					`CREATE TABLE IF NOT EXISTS member_relations (
							member_id   UUID NOT NULL,
							org_id      UUID NOT NULL,
							role        VARCHAR(10) NOT NULL,
							created_at  TIMESTAMPTZ,
							updated_at  TIMESTAMPTZ,
							FOREIGN KEY (org_id) REFERENCES orgs (id),
							PRIMARY KEY (member_id, org_id)
						 )`,
				},
				Down: []string{
					`DROP TABLE IF EXISTS orgs`,
					`DROP TABLE IF EXISTS member_relations`,
				},
			},
			{
				Id: "auth_3",
				Up: []string{
					`CREATE TABLE IF NOT EXISTS group_relations (
							group_id    UUID UNIQUE NOT NULL,
							org_id      UUID NOT NULL,
							created_at  TIMESTAMPTZ,
							updated_at  TIMESTAMPTZ,
							FOREIGN KEY (org_id) REFERENCES orgs (id),
							PRIMARY KEY (group_id, org_id)
						 )`,
				},
				Down: []string{
					`DROP TABLE IF EXISTS group_relations`,
				},
			},
			{
				Id: "auth_4",
				Up: []string{
					`CREATE TABLE IF NOT EXISTS users_roles (
					       	        role VARCHAR(12) CHECK (role IN ('root', 'admin')),
 				           	        user_id UUID NOT NULL,
				                        PRIMARY KEY (user_id)
				                 )`,
				},
				Down: []string{
					"DROP TABLE users_roles",
				},
			},
			{
				Id: "auth_5",
				Up: []string{
					`CREATE TABLE IF NOT EXISTS group_policies (
							group_id    UUID UNIQUE NOT NULL,
							member_id   UUID NOT NULL,
							policy      VARCHAR(15),
							FOREIGN KEY (group_id) REFERENCES group_relations (group_id) ON DELETE CASCADE ON UPDATE CASCADE
						 )`,
				},
				Down: []string{
					`DROP TABLE IF EXISTS group_policies`,
				},
			},
			{
				Id: "auth_6",
				Up: []string{
					`ALTER TABLE group_policies DROP CONSTRAINT IF EXISTS group_policies_group_id_key`,
					`ALTER TABLE group_policies ADD CONSTRAINT group_policies_pkey PRIMARY KEY (group_id, member_id)`,
				},
			},
			{
				Id: "auth_7",
				Up: []string{
					`ALTER TABLE member_relations DROP CONSTRAINT IF EXISTS member_relations_org_id_fkey`,
					`ALTER TABLE member_relations ADD CONSTRAINT member_relations_org_id_fkey FOREIGN KEY (org_id) REFERENCES orgs (id) ON DELETE CASCADE ON UPDATE CASCADE`,
					`ALTER TABLE group_relations DROP CONSTRAINT IF EXISTS group_relations_org_id_fkey`,
					`ALTER TABLE group_relations ADD CONSTRAINT group_relations_org_id_fkey FOREIGN KEY (org_id) REFERENCES orgs (id) ON DELETE CASCADE ON UPDATE CASCADE`,
				},
			},
		},
	}

	_, err := migrate.Exec(db.DB, "postgres", migrations, migrate.Up)
	return err
}
