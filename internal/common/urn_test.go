package common_test

import (
	"testing"

	"github.com/griffnb/core/lib/model"
	"github.com/griffnb/techboss-ai-go/internal/common"
)

func TestGenerateURN(_ *testing.T) {
	obj := &model.BaseModel{}
	obj.Initialize(&model.InitializeOptions{
		Table: "admins",
		Model: "Admin",
	})

	common.GenerateURN(obj)
	// log.PrintEntity(obj)
}
