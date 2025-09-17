package delay_queue

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/CrowdShield/go-core/lib/log"
	"github.com/griffnb/techboss-ai-go/internal/environment"
	"github.com/pkg/errors"
)

const TABLE_NAME = "task_delay_queue"

func AddDelayQueueTable() {
	// Builds the session tables
	tableInput := &dynamodb.CreateTableInput{
		TableName: aws.String(TABLE_NAME),
		KeySchema: []types.KeySchemaElement{
			{
				AttributeName: aws.String("id"),
				KeyType:       types.KeyTypeHash,
			},
		},
		AttributeDefinitions: []types.AttributeDefinition{
			{
				AttributeName: aws.String("id"),
				AttributeType: types.ScalarAttributeTypeS,
			},
			{
				AttributeName: aws.String("timestamp"),
				AttributeType: types.ScalarAttributeTypeN,
			},
			{
				AttributeName: aws.String("_static"),
				AttributeType: types.ScalarAttributeTypeN,
			},
		},
		GlobalSecondaryIndexes: []types.GlobalSecondaryIndex{
			{
				IndexName: aws.String("timestamp-index"),
				KeySchema: []types.KeySchemaElement{
					{
						AttributeName: aws.String("_static"),
						KeyType:       types.KeyTypeHash,
					},
					{
						AttributeName: aws.String("timestamp"),
						KeyType:       types.KeyTypeRange,
					},
				},
				Projection: &types.Projection{
					ProjectionType: types.ProjectionTypeAll,
				},
				ProvisionedThroughput: &types.ProvisionedThroughput{
					ReadCapacityUnits:  aws.Int64(10),
					WriteCapacityUnits: aws.Int64(10),
				},
			},
		},

		ProvisionedThroughput: &types.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(10),
			WriteCapacityUnits: aws.Int64(10),
		},
	}
	_, err := environment.GetDynamo().GetClient().CreateTable(context.TODO(), tableInput)
	// Table already exists is fine
	if err != nil {
		if !strings.Contains(err.Error(), "ResourceInUseException") && !strings.Contains(err.Error(), "Table already exists") &&
			!strings.Contains(err.Error(), "Cannot create preexisting table") {
			log.Error(errors.Wrapf(err, "Message Table Creation RawErr:%v", err.Error()))
		}
	}
}
