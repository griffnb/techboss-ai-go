package migrations

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	env "github.com/griffnb/core/lib/environment"
	senv "github.com/griffnb/techboss-ai-go/internal/environment"
	"github.com/griffnb/techboss-ai-go/internal/models/message"

	"github.com/griffnb/core/lib/log"
	"github.com/pkg/errors"
)

// BuildDynamo Builds the Dynamo Tables
func BuildDynamo() {
	_ = createDynamoSession() // Login Sessions
	message.AddMessageTable() // Message Table
}

func createDynamoSession() error {
	// Builds the session tables
	sessionTableInput := &dynamodb.CreateTableInput{
		TableName: aws.String(senv.GetConfig().Dynamo.SessionTable),
		KeySchema: []types.KeySchemaElement{
			{
				AttributeName: aws.String(senv.GetConfig().Dynamo.SessionKey),
				KeyType:       types.KeyTypeHash,
			},
		},
		AttributeDefinitions: []types.AttributeDefinition{
			{
				AttributeName: aws.String(senv.GetConfig().Dynamo.SessionKey),
				AttributeType: types.ScalarAttributeTypeS,
			},
		},

		ProvisionedThroughput: &types.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(10),
			WriteCapacityUnits: aws.Int64(10),
		},
	}
	_, cerr := env.Env().GetDynamo().GetClient().CreateTable(context.TODO(), sessionTableInput)
	// Table already exists is fine
	if cerr != nil {
		if !strings.Contains(cerr.Error(), "ResourceInUseException") && !strings.Contains(cerr.Error(), "Table already exists") &&
			!strings.Contains(cerr.Error(), "Cannot create preexisting table") {
			log.Error(errors.Wrapf(cerr, "Session Table Creation RawErr:%v", cerr.Error()))
			return errors.WithStack(cerr)
		}
	}
	return nil
}
