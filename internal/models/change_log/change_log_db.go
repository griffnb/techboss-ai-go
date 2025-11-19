package change_log

import (
	"context"
	"sync"
	"time"

	"github.com/griffnb/core/lib/log"
	"github.com/griffnb/core/lib/model"
	"github.com/griffnb/core/lib/model/coremodel"
	"github.com/griffnb/core/lib/model/fields"
	"github.com/griffnb/core/lib/tools"
	"github.com/griffnb/core/lib/tools/maps"
	"github.com/griffnb/core/lib/types"
	"github.com/griffnb/techboss-ai-go/internal/environment"
	"github.com/pkg/errors"
)

const TABLE string = "change_logs"

const (
	MODEL        string = "ChangeLog"
	PACKAGE      string = "change_log"
	IS_VERSIONED        = false
	CHANGE_LOGS         = true
	CLIENT              = environment.CLIENT_DEFAULT
)

var (
	registerOnce sync.Once
	Columns      *ChangeLog
)

func init() {
	RegisterFields()
	Columns = New()
}

func RegisterFields() {
	registerOnce.Do(func() {
		fields.RegisterFieldTypes(&Structure{})
	})
}

// For structure see model.ChangeLog

type Structure struct {
	JoinFields
	DBColumns
}
type DBColumns struct {
	ID_          *fields.UUIDField                   `column:"id"            type:"uuid"   pk:"true" default:"gen_random_uuid()"`
	ObjectID     *fields.UUIDField                   `column:"object_id"     type:"uuid"             default:"gen_random_uuid()" index:"true"`
	Type         *fields.StringField                 `column:"type"          type:"text"             default:""                  index:"true"`
	UserURN      *fields.StringField                 `column:"user_urn"      type:"text"             default:""                  index:"true"`
	ObjectURN    *fields.StringField                 `column:"object_urn"    type:"text"             default:""                  index:"true"`
	BeforeValues *fields.StructField[map[string]any] `column:"before_values" type:"jsonb"            default:"{}"`
	AfterValues  *fields.StructField[map[string]any] `column:"after_values"  type:"jsonb"            default:"{}"`
	Timestamp    *fields.IntField                    `column:"timestamp"     type:"bigint"           default:"0"                 index:"true"`
	CreatedAt    *fields.TimeField                   `column:"created_at"    type:"tswtz"            default:"CURRENT_TIMESTAMP" index:"true"`
	UpdatedAt    *fields.TimeField                   `column:"updated_at"    type:"tswtz"            default:"CURRENT_TIMESTAMP"`
}

type JoinFields struct {
	UserName *fields.StringField `column:"user_name" type:"text" default:""`
}

type ChangeLog struct {
	model.BaseModel
	DBColumns
}

type ChangeLogJoined struct {
	ChangeLog
	JoinFields
}

type initializable interface {
	InitializeWithChangeLogs(*model.InitializeOptions)
	Load(result map[string]any)
}

func New() *ChangeLog {
	return NewType[*ChangeLog]()
}

func NewType[T initializable]() T {
	obj := tools.NewObj[T]()
	obj.InitializeWithChangeLogs(&model.InitializeOptions{
		Table:       TABLE,
		Model:       MODEL,
		ChangeLogs:  CHANGE_LOGS,
		Package:     PACKAGE,
		IsVersioned: IS_VERSIONED,
	})
	err := fields.InitializeFields(obj)
	if err != nil {
		log.Error(err)
	}
	return obj
}

func load(result map[string]interface{}) *ChangeLog {
	obj := New()
	obj.Load(result)
	return obj
}

func (this *ChangeLog) beforeSave(ctx context.Context) {
	this.BaseBeforeSave(ctx)
}

func (this *ChangeLog) afterSave(ctx context.Context) {
	this.BaseAfterSave(ctx)
}

func (this *ChangeLog) Save(user coremodel.Model) error {
	return this.SaveWithContext(context.Background(), user)
}

func (this *ChangeLog) SaveWithContext(ctx context.Context, user coremodel.Model) error {
	this.beforeSave(ctx)
	_, _ = this.BaseSave(ctx, user)
	this.afterSave(ctx)
	return nil
}

/*
*
*	Finders
*
*
 */

func FindAll(ctx context.Context, options *model.Options) ([]*ChangeLog, error) {
	results, err := environment.GetDBClient(CLIENT).FindAll(ctx, TABLE, options)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	modelResults := make([]*ChangeLog, 0)
	for _, result := range results {
		obj := load(result)
		modelResults = append(modelResults, obj)
	}
	return modelResults, nil
}

func FindFirst(ctx context.Context, options *model.Options) (*ChangeLog, error) {
	result, err := environment.GetDBClient(CLIENT).FindFirst(ctx, TABLE, options)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return load(result), nil
}

func Get(ctx context.Context, id types.UUID) (*ChangeLog, error) {
	result, err := environment.GetDBClient(CLIENT).Find(ctx, TABLE, id)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return load(result), nil
}

func DBChangeLog(ctx context.Context, change *model.Change) {
	changeLog := New()
	changeLog.Timestamp.Set(time.Now().Unix())
	// Unfolds the map
	preData, err := maps.RecursiveJSON(change.PreMap)
	if err != nil {
		log.Error(errors.WithMessage(err, "pre data error"))
		return
	}

	// Unfolds the map
	afterData, err := maps.RecursiveJSON(change.SaveMap)
	if err != nil {
		log.Error(errors.WithMessage(err, "after data error"))
		return
	}

	changeLog.BeforeValues.Set(preData.(map[string]any))
	changeLog.AfterValues.Set(afterData.(map[string]any))
	changeLog.Type.Set(change.Table)
	if !tools.Empty(change.SavingUser) {
		changeLog.UserURN.Set(change.SavingUser.GetString("urn"))
	}
	changeLog.ObjectID.Set(change.ID)
	changeLog.ObjectURN.Set(change.URN)
	err = changeLog.SaveWithContext(ctx, change.SavingUser)
	if err != nil {
		log.Error(errors.WithMessagef(err, "saving change log"))
		return
	}
}

func FindResultsCount(ctx context.Context, options *model.Options) (int64, error) {
	return environment.GetDBClient(CLIENT).FindResultsCount(ctx, TABLE, options)
}
