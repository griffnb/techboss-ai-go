package object_tag

import (
	"context"
	"fmt"

	"github.com/CrowdShield/go-core/lib/model"
	"github.com/CrowdShield/go-core/lib/model/coremodel"
	"github.com/CrowdShield/go-core/lib/types"
	"github.com/griffnb/techboss-ai-go/internal/constants"
	"github.com/griffnb/techboss-ai-go/internal/environment"
)

type JoinedTag struct {
	TagID       types.UUID        `json:"tag_id"`
	Name        string            `json:"name"`
	Type        constants.TagType `json:"type"`
	Internal    bool              `json:"internal"`
	ObjectTagID types.UUID        `json:"object_tag_id"`
	ObjectURN   string            `json:"object_urn"`
}

func JoinQuery(targetTable string) string {
	return fmt.Sprintf(`
LEFT JOIN (
    SELECT
        ot.object_urn,
        json_agg(json_build_object(
            'object_tag_id', ot.id,
            'tag_id', t.id,
            'name', t.name,
            'type', t.type,
            'internal', t.internal,
			'object_urn', ot.object_urn
        ) ORDER BY t.id)::jsonb AS tags,
        json_agg(t.id ORDER BY t.id)::jsonb AS tag_ids
    FROM object_tags ot
    JOIN tags t ON t.id = ot.tag_id
    WHERE t.disabled = 0
    GROUP BY ot.object_urn
) tag_data ON tag_data.object_urn = %s.urn
	`, targetTable)
}

func PublicJoinQuery(targetTable string) string {
	return fmt.Sprintf(`
LEFT JOIN (
	SELECT 
		object_tags.object_urn, 
		json_agg(
			json_build_object(
				'object_tag_id', object_tags.id, 
				'tag_id', tags.id, 
				'name', tags.name, 
				'type', tags.type,
				'internal', tags.internal
			) ORDER BY tags.name
		)::jsonb as tags,
		json_agg(tags.id ORDER BY tags.name)::jsonb as tag_ids
	FROM object_tags
	JOIN tags ON tags.id = object_tags.tag_id
	WHERE tags.disabled = 0 AND tags.internal = 0
	GROUP BY object_tags.object_urn
) tag_data ON tag_data.object_urn = %s.urn
	
	`, targetTable)
}

func JoinField() string {
	return "tag_data.tags::jsonb as tags"
}

func AddTagFilter(value any, options *model.Options, key ...string) *model.Options {
	if len(key) > 0 {
		return options.WithArrayFilter(value, "tag_data.tag_ids", key[0])
	}
	return options.WithArrayFilter(value, "tag_data.tag_ids", "tag_ids")
}

func (this *ObjectTag) Delete() error {
	err := environment.DB().DB.Insert("DELETE FROM object_tags WHERE id = :id:", map[string]any{
		":id:": this.ID(),
	})
	if err != nil {
		return err
	}
	return nil
}

func AddTag(objectURN string, tagID types.UUID, savingUser coremodel.Model) (*ObjectTag, error) {
	obj := New()
	obj.ObjectURN.Set(objectURN)
	obj.TagID.Set(tagID)
	err := obj.SaveIfNotExists(savingUser)
	if err != nil {
		return nil, err
	}
	return obj, nil
}

// AddExclusiveTag adds a tag to an object, removing any existing tags of the same type
func AddExclusiveTag(
	ctx context.Context,
	tagType constants.TagType,
	objectURN string,
	tagID types.UUID,
	savingUser coremodel.Model,
) (*ObjectTag, error) {
	existingTags, err := GetByObjectAndType(ctx, objectURN, tagType)
	if err != nil {
		return nil, err
	}

	for _, existingTag := range existingTags {
		if existingTag.TagID.Get() == tagID {
			return existingTag, nil
		}
	}

	for _, existingTag := range existingTags {
		err = existingTag.Delete()
		if err != nil {
			return nil, err
		}
	}

	obj := New()
	obj.ObjectURN.Set(objectURN)
	obj.TagID.Set(tagID)
	err = obj.SaveIfNotExists(savingUser)
	if err != nil {
		return nil, err
	}
	return obj, nil
}

func GetByObjectAndType(ctx context.Context, objectURN string, tagType constants.TagType) ([]*ObjectTag, error) {
	return FindAll(ctx, &model.Options{
		Conditions: fmt.Sprintf("%s = :object_urn: AND tags.type = :type:", Columns.ObjectURN.Column()),
		Joins:      []string{"JOIN tags ON tags.id = object_tags.tag_id"},
		Params: map[string]interface{}{
			":object_urn:": objectURN,
			":type:":       tagType,
		},
	})
}

func GetByObjectAndTag(ctx context.Context, objectURN string, tagID types.UUID) (*ObjectTag, error) {
	return FindFirst(ctx, &model.Options{
		Conditions: fmt.Sprintf("%s = :object_urn: AND %v = :tag_id:", Columns.ObjectURN.Column(), Columns.TagID.Column()),
		Params: map[string]interface{}{
			":object_urn:": objectURN,
			":tag_id:":     tagID,
		},
	})
}

func GetByObject(ctx context.Context, objectURN string) ([]*ObjectTag, error) {
	return FindAll(ctx,
		model.NewOptions().
			WithCondition("%s = :object_urn:", Columns.ObjectURN.Column()).
			WithParam(":object_urn:", objectURN))
}
