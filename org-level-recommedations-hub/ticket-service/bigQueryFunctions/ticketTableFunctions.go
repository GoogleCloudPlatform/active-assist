package bigQueryFunctions

import (
	"context"
	"fmt"

	"cloud.google.com/go/bigquery"
)

var schema = bigquery.Schema{
	{Name: "IssueKey", Type: bigquery.IntegerFieldType},
	{Name: "CreationDate", Type: bigquery.TimestampFieldType},
	{Name: "Status", Type: bigquery.StringFieldType},
	{Name: "TargetResource", Type: bigquery.StringFieldType},
	{Name: "RecommenderIDs", Type: bigquery.StringFieldType, Repeated: true},
	{Name: "LastUpdateDate", Type: bigquery.TimestampFieldType},
	{Name: "LastPingDate", Type: bigquery.TimestampFieldType},
	{Name: "SnoozeDate", Type: bigquery.TimestampFieldType},
	{Name: "Subject", Type: bigquery.StringFieldType},
	{Name: "Assignee", Type: bigquery.StringFieldType},
}

// Example to use below
//ctx := context.Background()
//projectID := "my-project-id"
//datasetID := "my-dataset"
//tableID := "my-table"
//if err := createTable(ctx, projectID, datasetID, tableID); err != nil {
// Handle error
//}

func createTable(ctx context.Context, projectID string, datasetID string, tableID string) error {
	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		return err
	}

	metadata := &bigquery.TableMetadata{
		Name:   tableID,
		Schema: schema,
	}

	tableRef := client.Dataset(datasetID).Table(tableID)

	if err := tableRef.Create(ctx, metadata); err != nil {
		return err
	}

	fmt.Printf("Table %s:%s.%s created successfully\n", projectID, datasetID, tableID)
	return nil
}

func updateTableSchema(ctx context.Context, projectID string, datasetID string, tableID string) error {
	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		return err
	}

	tableRef := client.Dataset(datasetID).Table(tableID)
	metadata, err := tableRef.Metadata(ctx)
	if err != nil {
		return err
	}
	update := bigquery.TableMetadataToUpdate{
		Schema: schema,
	}
	if _, err := tableRef.Update(ctx, update, metadata.ETag); err != nil {
		return err
	}

	fmt.Printf("Table %s:%s.%s schema updated successfully\n", projectID, datasetID, tableID)
	return nil
}

func appendRowsToTable(ctx context.Context, projectID string, datasetID string, tableID string, rows []*bigquery.StructSaver) error {
	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		return err
	}

	tableRef := client.Dataset(datasetID).Table(tableID)
	_, err = tableRef.Metadata(ctx)
	if err != nil {
		return err
	}

	inserter := tableRef.Inserter()
	if err := inserter.Put(ctx, rows); err != nil {
		return err
	}

	return nil
}

func UpsertTicket(ctx context.Context, bqClient *bigquery.Client, datasetID, tableID string, ticket Ticket) error {
	inserter := bqClient.Dataset(datasetID).Table(tableID).Inserter()

	// Put() will handle both inserts and updates based on the existence of the row
	if err := inserter.Put(ctx, &ticket); err != nil {
		return fmt.Errorf("failed to insert/update ticket: %v", err)
	}
	return nil
}
