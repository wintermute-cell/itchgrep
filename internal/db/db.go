package db

import (
	"context"
	"errors"
	"itchgrep/internal/logging"
	"itchgrep/pkg/models"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
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
		dbEndpoint := "http://dynamodb-local:8000"
		if os.Getenv("DOCKER") != "true" {
			dbEndpoint = "http://localhost:8000"
		}
		cfg, err = config.LoadDefaultConfig(context.TODO(),
			config.WithCredentialsProvider(credentials.StaticCredentialsProvider{
				Value: aws.Credentials{
					AccessKeyID: "dummy", SecretAccessKey: "dummy", SessionToken: "dummy",
					Source: "Hard-coded credentials; values are irrelevant for local DynamoDB",
				},
			}),
			config.WithEndpointResolverWithOptions(aws.EndpointResolverWithOptionsFunc(
				func(service, region string, options ...interface{}) (aws.Endpoint, error) {
					if service == dynamodb.ServiceID {
						return aws.Endpoint{
							PartitionID:   "aws",
							URL:           dbEndpoint, // DynamoDB Local endpoint
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

// PutAssets puts a slice of assets into the database in chunks of 10 per transaction.
func PutAssets(svc *dynamodb.Client, assets []models.Asset) error {
	chunks := chunkAssets(assets, 25) // Chunk size of 10

	progress := 0
	for _, chunk := range chunks {
		logging.Info("Putting chunk %d of %d", progress, len(chunks))
		progress += 1
		var transactItems []types.TransactWriteItem
		for _, asset := range chunk {
			av, err := attributevalue.MarshalMap(asset)
			if err != nil {
				return err
			}

			transactItem := types.TransactWriteItem{
				Put: &types.Put{
					TableName: aws.String("Assets"),
					Item:      av,
				},
			}

			transactItems = append(transactItems, transactItem)
		}

		// Execute the transaction for the current chunk
		_, err := svc.TransactWriteItems(context.TODO(), &dynamodb.TransactWriteItemsInput{
			TransactItems: transactItems,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

// chunkAssets splits the slice of assets into chunks of a specified size.
func chunkAssets(assets []models.Asset, chunkSize int) [][]models.Asset {
	var chunks [][]models.Asset
	for i := 0; i < len(assets); i += chunkSize {
		end := i + chunkSize
		if end > len(assets) {
			end = len(assets)
		}
		chunks = append(chunks, assets[i:end])
	}
	return chunks
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

func GetAllAssets(svc *dynamodb.Client) ([]models.Asset, error) {
	result, err := svc.Scan(context.TODO(), &dynamodb.ScanInput{
		TableName: aws.String("Assets"),
	})

	if err != nil {
		return nil, err
	}

	var assets []models.Asset
	for _, i := range result.Items {
		var asset models.Asset
		err = attributevalue.UnmarshalMap(i, &asset)
		if err != nil {
			return nil, err
		}
		assets = append(assets, asset)
	}

	return assets, nil
}
