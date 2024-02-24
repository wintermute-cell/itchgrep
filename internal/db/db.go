package db

import (
	"context"
	"errors"
	"itchgrep/internal/logging"
	"itchgrep/pkg/models"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

func CreateDynamoClient(useLocalDynamo bool) (*dynamodb.Client, error) {
	var cfg aws.Config
	var err error
	if !useLocalDynamo {
		cfg, err = config.LoadDefaultConfig(
			context.TODO(),
			config.WithRegion("eu-central-1"))
		if err != nil {
			return nil, err
		}
	} else {
		logging.Info("Using local DynamoDB")
		cfg, err = config.LoadDefaultConfig(context.TODO(),
			config.WithEndpointResolverWithOptions(aws.EndpointResolverWithOptionsFunc(
				func(service, region string, options ...interface{}) (aws.Endpoint, error) {
					if service == dynamodb.ServiceID {
						return aws.Endpoint{
							PartitionID:   "aws",
							URL:           "http://localhost:8000", // DynamoDB Local endpoint
							SigningRegion: "eu-central-1",
						}, nil
					}
					return aws.Endpoint{}, &aws.EndpointNotFoundError{}
				})),
			config.WithRegion("eu-central-1"),
		)
	}

	// Create an Amazon DynamoDB service client
	svc := dynamodb.NewFromConfig(cfg)
	return svc, nil
}

func CrateAssetsTableIfNotExists(svc *dynamodb.Client) error {
	tableName := "Assets"
	if checkTableExists(svc, tableName) {
		return nil
	}
	logging.Info("Creating table %s", tableName)
	_, err := svc.CreateTable(context.TODO(), &dynamodb.CreateTableInput{
		TableName: &tableName,
		KeySchema: []types.KeySchemaElement{
			{
				AttributeName: aws.String("GameId"),
				KeyType:       types.KeyTypeHash, // Partition key
			},
		},
		AttributeDefinitions: []types.AttributeDefinition{
			{
				AttributeName: aws.String("GameId"),
				AttributeType: types.ScalarAttributeTypeS,
			},
		},
		BillingMode: types.BillingModePayPerRequest,
	})
	return err
}

func checkTableExists(svc *dynamodb.Client, tableName string) bool {
	_, err := svc.DescribeTable(context.TODO(), &dynamodb.DescribeTableInput{
		TableName: aws.String(tableName),
	})

	// If there's no error, the table exists
	if err == nil {
		return true
	}

	// If the error is because the table does not exist, return false
	var notFound *types.ResourceNotFoundException
	if ok := errors.As(err, &notFound); ok {
		return false
	}

	// For any other error, panic
	panic("Failed to describe table: " + err.Error())
}

func PutAsset(svc *dynamodb.Client, asset models.Asset) error {
	logging.Info("Putting asset with id: %s", asset.GameId)
	av, err := attributevalue.MarshalMap(asset)
	if err != nil {
		return err
	}

	_, err = svc.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: aws.String("Assets"),
		Item:      av,
	})

	if err != nil {
		return err
	}
	return nil
}

func GetAsset(svc *dynamodb.Client, gameId string) (models.Asset, error) {
	result, err := svc.GetItem(context.TODO(), &dynamodb.GetItemInput{
		TableName: aws.String("Assets"),
		Key: map[string]types.AttributeValue{
			"GameId": &types.AttributeValueMemberS{Value: gameId},
		},
	})

	if err != nil {
		return models.Asset{}, err
	}

	var assetRead models.Asset
	err = attributevalue.UnmarshalMap(result.Item, &assetRead)
	if err != nil {
		return models.Asset{}, err
	}

	return assetRead, nil
}
