package migrations

import "github.com/CrowdShield/go-core/lib/model"

func init() {
	model.AddMigration(&model.Migration{
		ID:           4,
		Table:        "",
		SQLMigration: `CREATE EXTENSION IF NOT EXISTS pg_trgm;`,
	})
}
