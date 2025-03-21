package main

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"log"
	"os"
)

func printItems(items []RentalItem) {
	fmt.Printf("%-4s | %-20s | %-15s | %6s | %s\n", "ID", "Name", "Warehouse", "Qty", "Price")
	fmt.Println("------------------------------------------------------------")
	for _, item := range items {
		fmt.Printf("%-4s | %-20s | %-15s | %6d | %d\n",
			item.ID, item.Name, item.Warehouse, item.Quantity, item.Price)
	}
}

func main() {
	ctx := context.Background()

	cfg, err := config.LoadDefaultConfig(
		context.TODO(),
		config.WithRegion(os.Getenv("AWS_REGION")),
		config.WithBaseEndpoint(os.Getenv("DYNAMO_ENDPOINT")),
	)
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	client := dynamodb.NewFromConfig(cfg)
	manager := NewTableManager(client)

	if err := manager.CreateTable(ctx); err != nil {
		log.Fatal(err)
	}

	if err := manager.PopulateTable(ctx); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Initial data:")
	if items, err := manager.GetItems(ctx); err != nil {
		log.Fatal(err)
	} else {
		printItems(items)
	}

	if err := manager.UpdateQuantity(ctx, "001", 5); err != nil {
		log.Fatal(err)
	}

	if err := manager.DeleteItem(ctx, "007"); err != nil {
		log.Fatal(err)
	}

	fmt.Println("\nAfter updates:")
	if items, err := manager.GetItems(ctx); err != nil {
		log.Fatal(err)
	} else {
		printItems(items)
	}

	if err := manager.ClearTable(ctx); err != nil {
		log.Fatal(err)
	}

	fmt.Println("\nAfter clearing:")
	if items, err := manager.GetItems(ctx); err != nil {
		log.Fatal(err)
	} else {
		printItems(items)
	}
}
