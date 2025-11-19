package change_log

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/griffnb/core/lib/log"
	"github.com/griffnb/techboss-ai-go/internal/environment"
	"github.com/pkg/errors"
)

func AddChangeLogTable() {
	// Builds the tables
	createTableInput := &dynamodb.CreateTableInput{
		TableName: aws.String(TABLE),
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
				AttributeName: aws.String("user_urn"),
				AttributeType: types.ScalarAttributeTypeS,
			},
			{
				AttributeName: aws.String("object_urn"),
				AttributeType: types.ScalarAttributeTypeS,
			},
		},
		GlobalSecondaryIndexes: []types.GlobalSecondaryIndex{
			{
				IndexName: aws.String("object_urn-index"),
				KeySchema: []types.KeySchemaElement{
					{
						AttributeName: aws.String("object_urn"),
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
					WriteCapacityUnits: aws.Int64(20),
				},
			},
			{
				IndexName: aws.String("user_urn-index"),
				KeySchema: []types.KeySchemaElement{
					{
						AttributeName: aws.String("user_urn"),
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
					WriteCapacityUnits: aws.Int64(20),
				},
			},
		},

		ProvisionedThroughput: &types.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(10),
			WriteCapacityUnits: aws.Int64(20),
		},
	}
	_, err := environment.GetDynamo().GetClient().CreateTable(context.TODO(), createTableInput)
	// Table already exists is fine
	if err != nil {
		if !strings.Contains(err.Error(), "ResourceInUseException") && !strings.Contains(err.Error(), "Table already exists") &&
			!strings.Contains(err.Error(), "Cannot create preexisting table") {
			log.Error(errors.Wrapf(err, "Message Table Creation RawErr:%v", err.Error()))
		}
	}
}
