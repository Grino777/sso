package migrations

import "embed"

//go:embed migrations/*.sql
var EmbedMigrations embed.FS

func GetEmbedFS() embed.FS {
	return EmbedMigrations
}
