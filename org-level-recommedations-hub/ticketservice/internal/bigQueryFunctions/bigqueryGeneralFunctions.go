package bigqueryfunctions

import (
	"context"
	"fmt"
	"log"

	"cloud.google.com/go/bigquery"
	"google.golang.org/api/iterator"
)

func QueryBigQuery(projectID string, query string) ([][]bigquery.Value, error) {
	ctx := context.Background()

	// Create a BigQuery client
	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	q := client.Query(query)

	// Run the query
	job, err := q.Run(ctx)
	if err != nil {
		fmt.Printf("Failed to run query: %v", err)
		return nil, err
	}

	// Wait for the query to complete
	status, err := job.Wait(ctx)
	if err != nil {
		fmt.Printf("Failed to wait for job completion: %v", err)
		return nil, err
	}
	if err := status.Err(); err != nil {
		fmt.Printf("Query error: %v", err)
		return nil, err
	}

	// Get the results
	iter, err := job.Read(ctx)
	if err != nil {
		fmt.Printf("Failed to read results: %v", err)
		return nil, err
	}
	var results [][]bigquery.Value
	for {
		var row []bigquery.Value
		err := iter.Next(&row)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		results = append(results, row)
	}

	return results, nil
}