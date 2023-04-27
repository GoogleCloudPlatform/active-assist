package bigqueryfunctions

import (
	"cloud.google.com/go/bigquery"
	"fmt"
	"reflect"
)

type routingRow struct {
	Target string
	ProjectID string
	TicketSystemIdentifiers	[]string
}

var routingSchema = bigquery.Schema{
	{Name: "Target", Type: bigquery.StringFieldType, Required: true},
	{Name: "ProjectID", Type: bigquery.StringFieldType},
	{Name: "TicketSystemIdentifiers", Type: bigquery.StringFieldType, Repeated: true},
}

var getTargetByProjectIDQuery = `Select * from %v.%v.%v where ProjectID = "%v" limit 1`

func GetRoutingRowsByProjectID(tableID string, project string)([]routingRow, error){
	query := fmt.Sprintf(getTargetByProjectIDQuery,projectID, datasetID, tableID, project)
	t := reflect.TypeOf(routingRow{})
	results, err := QueryBigQueryToStruct(query, t)
	if err != nil {
		return nil, err
	}
	// Type assertion to convert results to []routingRow
	var rows []routingRow
	for _, row := range results {
		if r, ok := row.(routingRow); ok {
			rows = append(rows, r)
		} else {
			// Handle type assertion error
			return nil, fmt.Errorf("failed to assert type routingRow")
		}
	}

	return rows, nil
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