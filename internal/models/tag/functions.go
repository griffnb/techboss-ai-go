package tag

import (
	"context"
	"fmt"
	"strings"

	"github.com/griffnb/core/lib/model"
	"github.com/griffnb/core/lib/model/coremodel"
	"github.com/griffnb/core/lib/tools"
	"github.com/griffnb/techboss-ai-go/internal/constants"
)

func GetByNameAndType(ctx context.Context, name string, tagType constants.TagType) (*Tag, error) {
	return FindFirst(ctx, &model.Options{
		Conditions: fmt.Sprintf("lower(%s) = :name: AND %s = :type:", Columns.Name.Column(), Columns.Type.Column()),
		Params: map[string]interface{}{
			":name:": strings.ToLower(name),
			":type:": tagType,
		},
	})
}

func GetOrCreateByNameAndType(ctx context.Context, name string, tagType constants.TagType, savingUser coremodel.Model) (*Tag, error) {
	existingTag, err := GetByNameAndType(ctx, name, tagType)
	if err != nil {
		return nil, err
	}

	if tools.Empty(existingTag) {
		tagObj := New()
		tagObj.Name.Set(name)
		tagObj.Type.Set(tagType)
		err = tagObj.Save(savingUser)
		if err != nil {
			return nil, err
		}
		existingTag = tagObj
	}

	return existingTag, nil
}
