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

package main

import (
	"fmt"
	"regexp"
	"strings"
	"reflect"
	b "ticketservice/internal/bigqueryfunctions"
	"ticketservice/internal/ticketinterfaces"
	u "ticketservice/internal/utils"
	"time"

)


var nonAlphanumericRegex = regexp.MustCompile(`[^a-zA-Z0-9 ]+`)

// %[1] is the recommender export table
// %[2] is the ticket table
// %[3] is the Cost Threshold
// %[4] is an additional string added to allow null values
// TODO: (GHAUN) reduce the number of returned fields
var checkQueryTpl = `SELECT f.* EXCEPT(
	recommender_last_refresh_time,
	has_impact_cost,
	recommender_state,
	folder_ids,
	insights,
	insight_ids,
	target_resources),
	TargetResource,
	struct(
			IFNULL(t.IssueKey, "") as IssueKey,
			IFNULL(t.TargetContact, "") as TargetContact,
			IFNULL(t.CreationDate, TIMESTAMP '1970-01-01T00:00:00Z') as CreationDate,
			IFNULL(t.Status, "") as Status,
			IFNULL(t.TargetResource, "") as TargetResource,
			IFNULL(t.RecommenderID, "") as RecommenderID,
			IFNULL(t.LastUpdateDate, TIMESTAMP '1970-01-01T00:00:00Z') as LastUpdateDate,
			IFNULL(t.LastPingDate, TIMESTAMP '1970-01-01T00:00:00Z') as LastPingDate,
			IFNULL(t.SnoozeDate, TIMESTAMP '1970-01-01T00:00:00Z') as SnoozeDate,
			IFNULL(t.Subject, "") as Subject,
			t.Assignee
			) as Ticket
	FROM %[1]s as f 
	cross join unnest(target_resources) as TargetResource 
	Left Join %[2]s as t 
	on TargetResource=t.TargetResource 
	where (t.IssueKey IS NULL or CURRENT_TIMESTAMP() >= SnoozeDate) and
	(impact_cost_unit >= %[3]d %[4]s) 
	and recommender_subtype not in (%[5]s)
	limit 1` // This is temporary.


func checkAndCreateNewTickets() error {
	var allowNullString string
	if c.AllowNullCost {
		allowNullString = "or impact_cost_unit is null"
	}
	query := fmt.Sprintf(checkQueryTpl, 
		fmt.Sprintf("%s.%s", c.BqDataset, c.BqRecommendationsTable),
		fmt.Sprintf("%s.%s", c.BqDataset, c.BqTicketTable),
		c.TicketCostThreshold,
		allowNullString,
		c.ExcludeSubTypes,
	)
	u.LogPrint(1, "Querying for new Tickets")
	t := reflect.TypeOf(ticketinterfaces.RecommendationQueryResult{})
	results, err := b.QueryBigQueryToStruct(query, t)
	if err != nil {
		u.LogPrint(4,"Failed to query bigquery for new tickets")
		return err
	}
	var rowsToInsert []ticketinterfaces.Ticket
	for _, r := range results{
		row, ok := r.(ticketinterfaces.RecommendationQueryResult);
		if !ok {
			return fmt.Errorf("Failed to convert Query Schema into RecommendationQueryResults")
		}
		ticket := row.Ticket
		// Logic for if the ticket is already created
		if ticket.IssueKey != ""{
			u.LogPrint(3,"Already Exists: " + ticket.IssueKey)
			ticket.RecommenderID = row.Recommender_name
			ticket.SnoozeDate = time.Now().AddDate(0,0,7)
			rowsToInsert = append(rowsToInsert, ticket)
			continue;
		}
		u.LogPrint(1, "Retrieving Routing Information")
		routingRows, err := b.GetRoutingRowsByProjectID(c.BqRoutingTable,row.Project_id)
		if err != nil {
			u.LogPrint(3,"Failed to get routing information")
			return err
		}
		u.LogPrint(1,"Creating new Ticket")
		ticket.TargetContact = routingRows[0].Target
		// And then adding to BQ Table.
		lastSlashIndex := strings.LastIndex(row.TargetResource, "/")
		secondToLast := strings.LastIndex(row.TargetResource[:lastSlashIndex], "/")
		// Update the fields of the ticket that need updating from the map
		ticket.CreationDate = time.Now()
		ticket.LastUpdateDate = time.Now()
		ticket.LastPingDate = time.Now()
		ticket.SnoozeDate = time.Now().AddDate(0,0,7)
		// For the subject we need to remove all special chars
		// One could argue this should be done in the Ticket Interface
		// We also need to combine target resource with recommender subtype
		// This may not be the best format....but it works for now
		ticket.Subject = fmt.Sprintf("%s-%s",
				row.Recommender_subtype,
				nonAlphanumericRegex.ReplaceAllString(
					row.TargetResource[secondToLast+1:],
					""))
		ticket.Assignee = routingRows[0].TicketSystemIdentifiers
		ticket.RecommenderID = row.Recommender_name
		// I need a way to catch IF a ticket is already created
		ticketID, err := ticketService.CreateTicket(ticket)
		if err != nil {
			return err
		}
		ticket.IssueKey = ticketID
		rowsToInsert = append(rowsToInsert, ticket)
	}
	err = b.AppendTicketsToTable(c.BqTicketTable, rowsToInsert)
	if err != nil {
		u.LogPrint(3,err)
		return err
	}
	return err
}