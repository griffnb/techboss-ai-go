package migrations

import (
	"github.com/griffnb/core/lib/model"
)

const TABLE string = "object_tags"

func init() {
	model.AddMigration(&model.Migration{
		ID:    1730859589,
		Table: TABLE,
		SQLMigration: `
		CREATE TABLE object_tags (
			"id" UUid DEFAULT gen_random_uuid() NOT NULL,
    		"object_urn" TEXT NOT NULL,
    		"tag_id" UUID REFERENCES tags(id) ON DELETE CASCADE,
			"status" SmallInt DEFAULT (0)::smallint NOT NULL,
			"created_by_urn" Text,
			"updated_by_urn" Text,
			"created_at" Timestamp With Time Zone DEFAULT CURRENT_TIMESTAMP NOT NULL,	
			"updated_at" Timestamp With Time Zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    		PRIMARY KEY (object_urn, tag_id)
		);

		CREATE INDEX idx_object_tags_id ON object_tags(id);
		CREATE INDEX idx_object_tags_created_at ON object_tags(created_at);
		CREATE INDEX idx_object_tags_updated_at ON object_tags(updated_at);
		CREATE INDEX idx_object_tags_created_by_urn ON object_tags(created_by_urn);
		CREATE INDEX idx_object_tags_updated_by_urn ON object_tags(updated_by_urn);
		CREATE INDEX idx_object_tags_status ON object_tags(status);

		`,
	})
}
