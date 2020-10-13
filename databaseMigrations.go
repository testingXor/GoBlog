package main

import (
	"database/sql"

	"github.com/lopezator/migrator"
)

func migrateDb() error {
	startWritingToDb()
	defer finishWritingToDb()
	m, err := migrator.New(
		migrator.Migrations(
			&migrator.Migration{
				Name: "00001",
				Func: func(tx *sql.Tx) error {
					_, err := tx.Exec(`
					CREATE TABLE posts (path text not null primary key, content text, published text, updated text, blog text not null, section text);
					CREATE TABLE post_parameters (id integer primary key autoincrement, path text not null, parameter text not null, value text);
					CREATE INDEX index_pp_path on post_parameters (path);
					CREATE TABLE redirects (fromPath text not null, toPath text not null, primary key (fromPath, toPath));
					CREATE TABLE indieauthauth (time text not null, code text not null, me text not null, client text not null, redirect text not null, scope text not null);
					CREATE TABLE indieauthtoken (time text not null, token text not null, me text not null, client text not null, scope text not null);
					CREATE INDEX index_iat_token on indieauthtoken (token);
					`)
					return err
				},
			},
		),
	)
	if err != nil {
		return err
	}
	if err := m.Migrate(appDb); err != nil {
		return err
	}
	return nil
}
