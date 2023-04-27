package bigqueryfunctions

import (
	"fmt"
	"strings"
	t "ticketservice/internal/ticketinterfaces"
	u "ticketservice/internal/utils"

	"cloud.google.com/go/bigquery"
)

var ticketSchema = bigquery.Schema{
	{Name: "IssueKey", Type: bigquery.StringFieldType, Required: true},
	{Name: "TargetContact", Type: bigquery.StringFieldType},
	{Name: "CreationDate", Type: bigquery.TimestampFieldType},
	{Name: "Status", Type: bigquery.StringFieldType},
	{Name: "TargetResource", Type: bigquery.StringFieldType},
	{Name: "RecommenderID", Type: bigquery.StringFieldType},
	{Name: "LastUpdateDate", Type: bigquery.TimestampFieldType},
	{Name: "LastPingDate", Type: bigquery.TimestampFieldType},
	{Name: "SnoozeDate", Type: bigquery.TimestampFieldType},
	{Name: "Subject", Type: bigquery.StringFieldType},
	{Name: "Assignee", Type: bigquery.StringFieldType, Repeated: true},
}

// An arguement could be made to make this a service that has it's own client.
// Will decide as I continue to develop

// createTable creates a BigQuery table in the specified dataset with the given table name and schema.
func createTicketTable(tableID string) error {

	if err := createTable(tableID, ticketSchema); err != nil{
		return err
	}
	// I couldn't find how to add this using GoLang library
	// Assuming since it's pre-ga it doesn't have it. 
	u.LogPrint(1,"Updating primary key")
	var addPrimaryKeyQuery = fmt.Sprintf(
		"ALTER TABLE `%s` ADD PRIMARY KEY (IssueKey) NOT ENFORCED",
		datasetID+"."+tableID,
	)
	_, err := QueryBigQueryToMap(addPrimaryKeyQuery)
	if err != nil {
		if !strings.Contains(err.Error(),"Already Exists"){
			return err
		}
	}

	// If the table was created successfully, log a message and return nil.
	u.LogPrint(1,"Table %s:%s.%s created successfully\n", client.Project(), datasetID, tableID)
	return nil
}

// CreateOrUpdateTable creates a BigQuery table or updates the schema if the table already exists.
// It takes a context, projectID, datasetID, and tableID as input.
// It returns an error if there is a problem creating or updating the table.
func CreateOrUpdateTicketTable(tableID string) error {
	// Create the table if it does not already exist.
	if err := createTicketTable(tableID); err != nil {
		return err
	}
	// Update the table schema if necessary.
	if err := updateTableSchema(tableID, ticketSchema); err != nil {
		return err
	}
	// Return nil if the table was created or updated successfully.
	return nil
}


// AppendTicketsToTable appends the provided tickets to a table in a BigQuery dataset.
// If the table does not exist, an error is returned.
func AppendTicketsToTable(tableID string, tickets []t.Ticket) error {
	// Get a reference to the target table.
	tableRef := client.Dataset(datasetID).Table(tableID)

	// Check if the target table exists.
	_, err := tableRef.Metadata(ctx)
	if err != nil {
		return err
	}

	// Create a new inserter for the target table.
	inserter := tableRef.Inserter()

	// Append the provided rows to the target table.
	if err := inserter.Put(ctx, tickets); err != nil {
		return err
	}
	u.LogPrint(1,"Inserted %d rows into BigQuery", len(tickets))
	return nil
}

// UpsertTicket inserts or updates a Ticket in a BigQuery table.
// The table must have a schema that matches the Ticket struct.
func UpsertTicket(tableID string, ticket t.Ticket) error {
	// Get a reference to the target table.
	tableRef := client.Dataset(datasetID).Table(tableID)

	// Create a new inserter for the target table.
	inserter := tableRef.Inserter()

	// Upsert the provided ticket into the target table.
	// The Put() method handles both inserts and updates based on the existence of the row.
	if err := inserter.Put(ctx, &ticket); err != nil {
		u.LogPrint(4,"failed to insert/update ticket: %v", err)
		return err
	}
	return nil
}
