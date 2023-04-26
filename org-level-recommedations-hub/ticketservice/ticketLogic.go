package main

import (
	"fmt"
	"regexp"
	"strings"
	bigqueryfunctions "ticketservice/internal/bigqueryfunctions"
	"ticketservice/internal/ticketinterfaces"
	u "ticketservice/internal/utils"
	"time"

	"github.com/mitchellh/mapstructure"
)

var nonAlphanumericRegex = regexp.MustCompile(`[^a-zA-Z0-9 ]+`)

// %[1] is the recommender export table
// %[2] is the ticket table
// %[3] is the Cost Threshold
// %[4] is an additional string added to allow null values
// TODO: (GHAUN) reduce the number of returned fields
var checkQueryTpl = `SELECT * EXCEPT(insights, insight_ids, target_resources)
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
	results, err := bigqueryfunctions.QueryBigQuery(query)
	if err != nil {
		u.LogPrint(4,"Failed to query bigquery for new tickets")
		return err
	}
	var rowsToInsert []ticketinterfaces.Ticket
	for _, row := range results{
		ticket := ticketinterfaces.Ticket{}
		if err := mapstructure.Decode(row, &ticket); err != nil{
			u.LogPrint(3,"Failed to decode row into ticket")
			return err
		}
		recommenderID := fmt.Sprintf("%s",row["recommender_name"])
		// Logic for if the ticket is already created
		if ticket.IssueKey != ""{
			u.LogPrint(3,"Already Exists: " + ticket.IssueKey)
			ticket.RecommenderID = recommenderID
			ticket.SnoozeDate = time.Now().AddDate(0,0,7)
			rowsToInsert = append(rowsToInsert, ticket)
			continue;
		}
		u.LogPrint(1,"Creating new Ticket")
		// Create ticket here
		// This involves creating the ticket in ticketInterface
		// And then adding to BQ Table.
		lastSlashIndex := strings.LastIndex(ticket.TargetResource, "/")
		secondToLast := strings.LastIndex(ticket.TargetResource[:lastSlashIndex], "/")
		// verify 
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
				row["recommender_subtype"],
				nonAlphanumericRegex.ReplaceAllString(
					ticket.TargetResource[secondToLast+1:],
					""))
		ticket.Assignee = "U03CS3FK54Z,U054RCYBMFA"
		ticket.RecommenderID = recommenderID
		// I need a way to catch IF a ticket is already created
		ticketID, err := ticketService.CreateTicket(ticket)
		if err != nil {
			return err
		}
		ticket.IssueKey = ticketID
		rowsToInsert = append(rowsToInsert, ticket)
	}
	err = bigqueryfunctions.AppendTicketsToTable(c.BqTicketTable, rowsToInsert)
	if err != nil {
		u.LogPrint(3,err)
		return err
	}
	return err
}