package main

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	bigqueryfunctions "ticketservice/internal/bigqueryfunctions"
	"ticketservice/internal/ticketinterfaces"
	"time"
)

var nonAlphanumericRegex = regexp.MustCompile(`[^a-zA-Z0-9 ]+`)

// %[1] is the recommender export table
// %[2] is the ticket table
// %[3] is the Cost Threshold
// %[4] is an additional string added to allow null values
// TODO: (GHAUN) reduce the number of returned fields
var checkQueryTpl = `SELECT * 
	FROM %[1]s as f 
	cross join unnest(target_resources) as target_resource 
	Left Join %[2]s as t 
	on target_resource=TargetResource 
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
	results, err := bigqueryfunctions.QueryBigQuery(
		c.BqProject, 
		query,
	)
	var rowsToInsert []ticketinterfaces.Ticket
	for _, row := range results{
		// Create ticket here
		// This involves creating the ticket in ticketInterface
		// And then adding to BQ Table.
		lastSlashIndex := strings.LastIndex(fmt.Sprintf("%v",row["target_resource"]), "/")
		ticket := ticketinterfaces.Ticket{
			IssueKey: "",
			CreationDate: time.Now(),
			Status: "Open",
			// Using Sprintf because it returns an int and we need to return string
			TargetResource: fmt.Sprintf("%v",row["target_resource"]),
			RecommenderIDs: []string{fmt.Sprintf("%v",row["recommender_name"])},
			LastUpdatedDate: time.Now(),
			LastPingDate: time.Now(),
			Subject: fmt.Sprintf("%s%s",
				nonAlphanumericRegex.ReplaceAllString(
					fmt.Sprintf("%s", row["target_resource"])[lastSlashIndex+1:],
					""),
				row["recommender_subtype"]),
			// Temp until I solve this
			Assignee: "thefsm93",
		}
		ticketID, err := ticketService.CreateTicket(ticket)
		if err != nil {
			return err
		}
		ticket.IssueKey = ticketID
		rowsToInsert = append(rowsToInsert, ticket)
	}
	err = bigqueryfunctions.AppendTicketsToTable(context.Background() ,c.BqProject, c.BqDataset, c.BqTicketTable, rowsToInsert)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return err
}