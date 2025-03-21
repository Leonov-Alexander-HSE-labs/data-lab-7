package main

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

const tableName = "RentalsTable"

type RentalItem struct {
	ID        string `dynamodbav:"id"`
	Name      string `dynamodbav:"name"`
	Warehouse string `dynamodbav:"warehouse"`
	Quantity  int    `dynamodbav:"quantity"`
	Price     int    `dynamodbav:"price"` // Бывшее поле SK
}

type TableManager struct {
	client *dynamodb.Client
}

func NewTableManager(client *dynamodb.Client) *TableManager {
	return &TableManager{
		client: client,
	}
}

func (tm *TableManager) CreateTable(ctx context.Context) error {
	_, err := tm.client.CreateTable(ctx, &dynamodb.CreateTableInput{
		TableName: aws.String(tableName),
		AttributeDefinitions: []types.AttributeDefinition{
			{AttributeName: aws.String("id"), AttributeType: types.ScalarAttributeTypeS},
		},
		KeySchema: []types.KeySchemaElement{
			{AttributeName: aws.String("id"), KeyType: types.KeyTypeHash},
		},
		ProvisionedThroughput: &types.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(5),
			WriteCapacityUnits: aws.Int64(5),
		},
	})

	return err
}

func (tm *TableManager) PopulateTable(ctx context.Context) error {
	items := []RentalItem{
		{"001", "Телевизор", "Нижегородский", 7, 10000},
		{"002", "Часы напольные", "Советский", 6, 5000},
		{"003", "Радиоприемник", "Нижегородский", 10, 7000},
		{"004", "Часы настенные", "Приокский", 20, 3000},
		{"005", "Холодильник", "Сормовский", 6, 12000},
		{"006", "Утюг", "Нижегородский", 30, 2000},
		{"007", "Весы детские", "Нижегородский", 15, 1500},
	}

	for _, item := range items {
		av, err := attributevalue.MarshalMap(item)
		if err != nil {
			return err
		}

		_, err = tm.client.PutItem(ctx, &dynamodb.PutItemInput{
			TableName: aws.String(tableName),
			Item:      av,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (tm *TableManager) GetItems(ctx context.Context) ([]RentalItem, error) {
	result, err := tm.client.Scan(ctx, &dynamodb.ScanInput{
		TableName: aws.String(tableName),
	})
	if err != nil {
		return nil, err
	}

	var items []RentalItem
	if err := attributevalue.UnmarshalListOfMaps(result.Items, &items); err != nil {
		return nil, err
	}

	return items, nil
}

func (tm *TableManager) UpdateQuantity(ctx context.Context, id string, newQty int) error {
	_, err := tm.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(tableName),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: id},
		},
		UpdateExpression: aws.String("SET quantity = :qty"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":qty": &types.AttributeValueMemberN{Value: fmt.Sprint(newQty)},
		},
	})
	return err
}

func (tm *TableManager) DeleteItem(ctx context.Context, id string) error {
	_, err := tm.client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(tableName),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: id},
		},
	})

	return err
}

func (tm *TableManager) ClearTable(ctx context.Context) error {
	result, err := tm.client.Scan(ctx, &dynamodb.ScanInput{
		TableName:            aws.String(tableName),
		ProjectionExpression: aws.String("id"),
	})
	if err != nil {
		return err
	}

	for _, item := range result.Items {
		id := item["id"].(*types.AttributeValueMemberS).Value
		if err := tm.DeleteItem(ctx, id); err != nil {
			return err
		}
	}
	return nil
}
