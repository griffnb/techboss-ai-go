package change_log

import (
	"context"
	"fmt"

	"github.com/CrowdShield/go-core/lib/model"
	"github.com/CrowdShield/go-core/lib/tools"
	"github.com/CrowdShield/go-core/lib/types"
	"github.com/griffnb/techboss-ai-go/internal/environment"
)

// AddJoinData adds in the join data
func AddJoinData(options *model.Options) {
	options.WithPrependJoins([]string{
		"LEFT JOIN accounts ON accounts.urn = change_logs.user_urn",
		"LEFT JOIN admins ON admins.urn = change_logs.user_urn",
	}...)
	options.WithIncludeFields([]string{
		"COALESCE(accounts.first_name, admins.name) AS user_name",
	}...)
}

// GetJoinedWithContext gets a record with a specific ID and joins the hierarchy to it
func GetJoinedWithContext(ctx context.Context, id types.UUID) (*ChangeLog, error) {
	options := &model.Options{
		Conditions: fmt.Sprintf("%s.id = :id:", TABLE),
		Params: map[string]interface{}{
			":id:": id,
		},
	}
	AddJoinData(options)
	return FindFirst(ctx, options)
}

// FindFirstJoinedWithContext Finds first record
func FindFirstJoinedWithContext(ctx context.Context, options *model.Options) (*ChangeLog, error) {
	AddJoinData(options)
	return FindFirst(ctx, options)
}

// FindAllJoinedWithContext Finds first record
func FindAllJoinedWithContext(ctx context.Context, options *model.Options) ([]*ChangeLog, error) {
	AddJoinData(options)
	return FindAll(ctx, options)
}

func GetCreators(urns []string) (map[string]string, error) {
	if tools.Empty(urns) {
		return map[string]string{}, nil
	}
	admins, err := environment.DB().DB.GetAll("SELECT urn, name FROM admins WHERE urn IN (:urns:)", map[string]any{
		":urns:": urns,
	})
	if err != nil {
		return nil, err
	}

	accounts, err := environment.DB().DB.GetAll(
		`SELECT urn, CONCAT(first_name,' ', last_name) as name FROM accounts WHERE urn IN (:urns:)`,
		map[string]any{
			":urns:": urns,
		},
	)
	if err != nil {
		return nil, err
	}

	creators := make(map[string]string)
	for _, admin := range admins {
		creators[admin["urn"].(string)] = admin["name"].(string)
	}

	for _, account := range accounts {
		creators[account["urn"].(string)] = account["name"].(string)
	}

	return creators, nil
}

func BuildChangeLogs(_ context.Context, logs []*ChangeLogDynamo) ([]*ChangeLogJoined, error) {
	urns := make([]string, 0)
	for _, log := range logs {
		urns = append(urns, log.UserURN)
	}

	creators, err := GetCreators(urns)
	if err != nil {
		return nil, err
	}

	changeLogs := make([]*ChangeLogJoined, 0)

	for _, log := range logs {
		changeLog := NewType[*ChangeLogJoined]()
		changeLog.ID_.Set(types.UUID(log.ID))
		changeLog.ObjectURN.Set(log.ObjectURN)
		changeLog.UserURN.Set(log.UserURN)
		changeLog.BeforeValues.Set(log.BeforeValues)
		changeLog.AfterValues.Set(log.AfterValues)
		changeLog.Timestamp.Set(log.Timestamp)

		if log.UserURN != "" {
			if usrName, ok := creators[log.UserURN]; ok {
				changeLog.UserName.Set(usrName)
			}
		}
		changeLogs = append(changeLogs, changeLog)
	}

	return changeLogs, nil
}
