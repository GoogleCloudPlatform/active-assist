package bigqueryfunctions

import (
	"context"
	"fmt"
	"strings"

	t "ticketservice/internal/ticketinterfaces"

	"cloud.google.com/go/bigquery"
)

var schema = bigquery.Schema{
	{Name: "IssueKey", Type: bigquery.StringFieldType, Required: true},
	{Name: "TargetContact", Type: bigquery.StringFieldType},
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

// An arguement could be made to make this a service that has it's own client.
// Will decide as I continue to develop

// createTable creates a BigQuery table in the specified dataset with the given table name and schema.
func createTable(ctx context.Context, client *bigquery.Client, projectID string, datasetID string, tableID string) error {

	// Define table metadata with table name and schema.
	metadata := &bigquery.TableMetadata{
		Name:   tableID,
		Schema: schema,
	}

	// Get a reference to the table using the datasetID and tableID.
	tableRef := client.Dataset(datasetID).Table(tableID)

	// Try to create the table with the given metadata.
	if err := tableRef.Create(ctx, metadata); err != nil {
		// If the table already exists, log a message and return nil.
		if strings.Contains(err.Error(), "Already Exists") {
			fmt.Printf("Table %s:%s.%s already exists\n", client.Project(), datasetID, tableID)
			return nil
		}
		// If there was an error creating the table that was not due to the table already existing, return the error.
		return err
	}
	// I couldn't find how to add this using GoLang library
	// Assuming since it's pre-ga it doesn't have it. 
	fmt.Println("Updating primary key")
	var addPrimaryKeyQuery = fmt.Sprintf(
		"ALTER TABLE `%s` ADD PRIMARY KEY (IssueKey) NOT ENFORCED",
		datasetID+"."+tableID,
	)
	_, err := QueryBigQuery(
		projectID, 
		addPrimaryKeyQuery,
	)
	if err != nil {
		return err
	}

	// If the table was created successfully, log a message and return nil.
	fmt.Printf("Table %s:%s.%s created successfully\n", client.Project(), datasetID, tableID)
	return nil
}

// updateTableSchema updates the schema of an existing BigQuery table
// with the given datasetID and tableID using the provided client.
func updateTableSchema(ctx context.Context, client *bigquery.Client, datasetID string, tableID string) error {
	// Get a reference to the table
	tableRef := client.Dataset(datasetID).Table(tableID)
	
	// Get the current metadata for the table
	metadata, err := tableRef.Metadata(ctx)
	if err != nil {
		return err
	}

	// Create an update object with the new schema
	update := bigquery.TableMetadataToUpdate{
		Schema: schema,
	}

	// Update the table with the new schema
	if _, err := tableRef.Update(ctx, update, metadata.ETag); err != nil {
		return err
	}

	// Print success message
	fmt.Printf("Table %s:%s.%s schema updated successfully\n", client.Project(), datasetID, tableID)
	return nil
}


// CreateOrUpdateTable creates a BigQuery table or updates the schema if the table already exists.
// It takes a context, projectID, datasetID, and tableID as input.
// It returns an error if there is a problem creating or updating the table.
func CreateOrUpdateTable(ctx context.Context, projectID string, datasetID string, tableID string) error {
	// Create a new BigQuery client.
	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		return err
	}
	// Create the table if it does not already exist.
	if err = createTable(ctx, client, projectID, datasetID, tableID); err != nil {
		return err
	}
	// Update the table schema if necessary.
	if err = updateTableSchema(ctx, client, datasetID, tableID); err != nil {
		return err
	}
	// Return nil if the table was created or updated successfully.
	return nil
}


// AppendTicketsToTable appends the provided tickets to a table in a BigQuery dataset.
// If the table does not exist, an error is returned.
func AppendTicketsToTable(ctx context.Context, projectID string, datasetID string, tableID string, tickets []t.Ticket) error {
	// Create a new BigQuery client using the provided project ID.
	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		return err
	}

	// Get a reference to the target table.
	tableRef := client.Dataset(datasetID).Table(tableID)

	// Check if the target table exists.
	_, err = tableRef.Metadata(ctx)
	if err != nil {
		return err
	}

	// Create a new inserter for the target table.
	inserter := tableRef.Inserter()

	// Append the provided rows to the target table.
	if err := inserter.Put(ctx, tickets); err != nil {
		return err
	}
	fmt.Printf("Inserted %d rows into BigQuery", len(tickets))
	return nil
}

// UpsertTicket inserts or updates a Ticket in a BigQuery table.
// The table must have a schema that matches the Ticket struct.
func UpsertTicket(ctx context.Context, bqClient *bigquery.Client, datasetID, tableID string, ticket t.Ticket) error {
	// Get a reference to the target table.
	tableRef := bqClient.Dataset(datasetID).Table(tableID)

	// Create a new inserter for the target table.
	inserter := tableRef.Inserter()

	// Upsert the provided ticket into the target table.
	// The Put() method handles both inserts and updates based on the existence of the row.
	if err := inserter.Put(ctx, &ticket); err != nil {
		return fmt.Errorf("failed to insert/update ticket: %v", err)
	}
	return nil
}
