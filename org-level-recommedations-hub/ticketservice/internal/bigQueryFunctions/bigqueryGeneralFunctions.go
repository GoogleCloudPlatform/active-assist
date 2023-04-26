package bigqueryfunctions

import (
	"context"
	"strings"

	u "ticketservice/internal/utils"

	"cloud.google.com/go/bigquery"
	"google.golang.org/api/iterator"
)

var (
	client *bigquery.Client
	projectID string
	datasetID string
	ctx context.Context
)


func InitBQ(dataset string, project string) error {
	datasetID = dataset
	projectID = project
	// Create a new BigQuery client.
	ctx = context.Background()
	bq, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		u.LogPrint(4,"Failed to create client: %v", err)
		return err
	}
	client = bq
	return nil
}

// QueryBigQuery executes the given BigQuery query and returns a map of field name to value for each row of the result.
func QueryBigQuery(query string) ([]map[string]interface{}, error) {

	q := client.Query(query)

	// Run the query
	job, err := q.Run(ctx)
	if err != nil {
		u.LogPrint(3,"Failed to run query: %v", err)
		return nil, err
	}

	// Wait for the query to complete
	status, err := job.Wait(ctx)
	if err != nil {
		u.LogPrint(3,"Failed to wait for job completion: %v", err)
		return nil, err
	}
	if err := status.Err(); err != nil {
		u.LogPrint(3,"Query error: %v", err)
		return nil, err
	}

	// Get the query results
	iter, err := job.Read(ctx)
	if err != nil {
		u.LogPrint(4,"Failed to read results: %v", err)
		return nil, err
	}

	// Get the results schema
	schema := iter.Schema

	var results []map[string]interface{}

	// Loop over the rows in the result
	for {
		var row []bigquery.Value
		err := iter.Next(&row)
		if err == iterator.Done {
			break
		}
		if err != nil {
			u.LogPrint(4,"Failed to read row: %v", err)
			return nil, err
		}

		// Create a map to hold the row data
		rowMap := make(map[string]interface{})

		// Loop over the fields in the row schema
		for i, field := range schema {
			// Add the field name and value to the row map
			rowMap[field.Name] = row[i]
		}

		// Add the row map to the results slice
		results = append(results, rowMap)
	}

	return results, nil
}

// createTable creates a BigQuery table in the specified dataset with the given table name and schema.
func createTable(tableID string, schema bigquery.Schema) error {

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
			u.LogPrint(3,"Table %s:%s.%s already exists\n", client.Project(), datasetID, tableID)
			return nil
		}
		// If there was an error creating the table that was not due to the table already existing, return the error.
		return err
	}
	// If the table was created successfully, log a message and return nil.
	u.LogPrint(1,"Table %s:%s.%s created successfully\n", client.Project(), datasetID, tableID)
	return nil
}

// updateTableSchema updates the schema of an existing BigQuery table
// with the given datasetID, tableID, and schema using the provided client.
func updateTableSchema(tableID string, schema bigquery.Schema) error {
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
	u.LogPrint(1,"Table %s:%s.%s schema updated successfully\n", client.Project(), datasetID, tableID)
	return nil
}