package bigqueryfunctions

import (
	"context"
	"fmt"

	"cloud.google.com/go/bigquery"
	"google.golang.org/api/iterator"
)

// QueryBigQuery executes the given BigQuery query and returns a map of field name to value for each row of the result.
func QueryBigQuery(projectID string, query string) ([]map[string]interface{}, error) {
	ctx := context.Background()

	// Create a BigQuery client
	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("Failed to create client: %v", err)
	}

	q := client.Query(query)

	// Run the query
	job, err := q.Run(ctx)
	if err != nil {
		return nil, fmt.Errorf("Failed to run query: %v", err)
	}

	// Wait for the query to complete
	status, err := job.Wait(ctx)
	if err != nil {
		return nil, fmt.Errorf("Failed to wait for job completion: %v", err)
	}
	if err := status.Err(); err != nil {
		return nil, fmt.Errorf("Query error: %v", err)
	}

	// Get the query results
	iter, err := job.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("Failed to read results: %v", err)
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
			return nil, fmt.Errorf("Failed to read row: %v", err)
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

