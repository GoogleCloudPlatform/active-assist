// Copyright 2020 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.```

package bigqueryfunctions

import (
	"fmt"
	"strings"
	t "ticketservice/internal/ticketinterfaces"
	u "ticketservice/internal/utils"
	"time"
	"reflect"
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

// This function is nuts, and probably way to complex for what it is.
// However, all the documentation and code I tried to make "upserting" work
// Using Primary Keys, etc in BQ have all failed. They all simply "appended" rows
// When in doubt, just write SQL. So that's where we are :/ 
func UpsertTicket(tableID string, ticket t.Ticket) error {
	// Get a reference to the target table.
	if tableID == "" {
		tableID = ticketTableID
	}
	// Build the update query.
	var updateStmts []string
	v := reflect.ValueOf(ticket)
	for i := 0; i < v.NumField(); i++ {
		field := v.Type().Field(i)
		fieldName := field.Name
		fieldValue := v.Field(i).Interface()

		// Convert the field value to a string representation
		var strValue string
		switch fieldValue := fieldValue.(type) {
		case []string:
			// Handle string arrays
			var strValues []string
			for _, val := range fieldValue {
				strValues = append(strValues, fmt.Sprintf("'%s'", val))
			}
			strValue = fmt.Sprintf("[%s]", strings.Join(strValues, ", "))
		case time.Time:
			// Handle time values
			strValue = "'" + fieldValue.Format("2006-01-02 15:04:05") + "'"
		case string:
			// Handle string values
			strValue = "'" + fieldValue + "'"
		default:
			strValue = fmt.Sprintf("%v", fieldValue)
		}

		// Skip fields with nil or empty values
		if fieldValue == nil || fieldValue == "" {
			continue
		}

		updateStmt := fmt.Sprintf("%s = %s", fieldName, strValue)
		updateStmts = append(updateStmts, updateStmt)
	}
	updateQuery := fmt.Sprintf("UPDATE `%s.%s` SET %s WHERE IssueKey = '%s'",
		datasetID, tableID, strings.Join(updateStmts, ", "), ticket.IssueKey)

	// Execute the update query.
	_, err := runQuery(updateQuery)
	if err != nil {
		u.LogPrint(4, "Failed to update ticket: %v", err)
		return err
	}

	return nil
}

func GetTicketByIssueKey(issueKey string) (*t.Ticket, error) {
	// Build the SQL query to retrieve the ticket with the matching issueKey.
	query := fmt.Sprintf("SELECT * FROM `%s.%s` WHERE IssueKey = '%s'", datasetID, ticketTableID, issueKey)
	tType := reflect.TypeOf(t.Ticket{})
	// Execute the query.
	ticket, err := QueryBigQueryToStruct(query, tType)
	if len(ticket) < 1 {
		u.LogPrint(3, "[TicketTableFunctions] Could not find ticket: %v", issueKey)
		return nil, fmt.Errorf("Could not find ticket: %v", issueKey)
	}
	if err != nil {
		u.LogPrint(3, "[TicketTableFunctions] Something went wrong querying ticket: %v", err)
		return nil, err
	}
	tick, ok := ticket[0].(t.Ticket);
	if !ok {
		u.LogPrint(3, "[TicketTableFunctions] Something went wrong asserting Ticket")
		return nil, fmt.Errorf("Assertion Error")
	} 
	return &tick, nil
}