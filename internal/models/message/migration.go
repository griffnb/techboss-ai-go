package message

import (
	"context"
	"strings"

	"github.com/CrowdShield/go-core/lib/log"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/griffnb/techboss-ai-go/internal/environment"
	"github.com/pkg/errors"
)

const TABLE_NAME = "chatbot_messages"

func AddMessageTable() {
	// Builds the session tables
	sessionTableInput := &dynamodb.CreateTableInput{
		TableName: aws.String(TABLE_NAME),
		KeySchema: []types.KeySchemaElement{
			{
				AttributeName: aws.String("key"),
				KeyType:       types.KeyTypeHash,
			},
		},
		AttributeDefinitions: []types.AttributeDefinition{
			{
				AttributeName: aws.String("key"),
				AttributeType: types.ScalarAttributeTypeS,
			},
			{
				AttributeName: aws.String("timestamp"),
				AttributeType: types.ScalarAttributeTypeN,
			},
			{
				AttributeName: aws.String("conversation_key"),
				AttributeType: types.ScalarAttributeTypeS,
			},
			{
				AttributeName: aws.String("contact_id"),
				AttributeType: types.ScalarAttributeTypeN,
			},
		},
		GlobalSecondaryIndexes: []types.GlobalSecondaryIndex{
			{
				IndexName: aws.String("conversation_key-index"),
				KeySchema: []types.KeySchemaElement{
					{
						AttributeName: aws.String("conversation_key"),
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
			{
				IndexName: aws.String("contact_id-index"),
				KeySchema: []types.KeySchemaElement{
					{
						AttributeName: aws.String("contact_id"),
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
	_, err := environment.GetDynamo().GetClient().CreateTable(context.TODO(), sessionTableInput)
	// Table already exists is fine
	if err != nil {
		if !strings.Contains(err.Error(), "ResourceInUseException") && !strings.Contains(err.Error(), "Table already exists") &&
			!strings.Contains(err.Error(), "Cannot create preexisting table") {
			log.Error(errors.Wrapf(err, "Message Table Creation RawErr:%v", err.Error()))
		}
	}
}
