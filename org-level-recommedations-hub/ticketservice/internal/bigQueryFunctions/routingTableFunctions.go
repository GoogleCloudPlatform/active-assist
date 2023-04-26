package bigqueryfunctions

import (
	"cloud.google.com/go/bigquery"
)

var routingSchema = bigquery.Schema{
	{Name: "Target", Type: bigquery.StringFieldType, Required: true},
	{Name: "ProjectID", Type: bigquery.StringFieldType},
	{Name: "TicketSystemIdentifiers", Type: bigquery.StringFieldType, Repeated: true},
}

func CreateOrUpdateRoutingTable(tableID string) error {
	// Create the table if it does not already exist.
	if err := createTable(tableID, routingSchema); err != nil {
		return err
	}
	// Update the table schema if necessary.
	if err := updateTableSchema(tableID, routingSchema); err != nil {
		return err
	}
	// Return nil if the table was created or updated successfully.
	return nil
}