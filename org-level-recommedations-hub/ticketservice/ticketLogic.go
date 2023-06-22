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
	"reflect"
	"sync"
	b "ticketservice/internal/bigqueryfunctions"
	"ticketservice/internal/ticketinterfaces"
	u "ticketservice/internal/utils"
	"time"

)



func checkAndCreateNewTickets() error {
	var allowNullString string
	if c.AllowNullCost {
		allowNullString = "or impact_cost_unit is null"
	}
	query := fmt.Sprintf(ticketinterfaces.CheckQueryTpl, 
		fmt.Sprintf("%s.%s", c.BqDataset, c.BqRecommendationsTable),
		fmt.Sprintf("%s.%s", c.BqDataset, c.BqTicketTable),
		c.TicketCostThreshold,
		allowNullString,
		c.ExcludeSubTypes,
		c.TicketLimitPerCall,
	)
	u.LogPrint(1, "Querying for new Tickets")
	t := reflect.TypeOf(ticketinterfaces.RecommendationQueryResult{})
	results, err := b.QueryBigQueryToStruct(query, t)
	if err != nil {
		u.LogPrint(4,"Failed to query bigquery for new tickets")
		return err
	}
	var rowsToInsert []ticketinterfaces.Ticket
	var rowsMutex sync.Mutex
	var wg sync.WaitGroup
	wg.Add(len(results))
	for _, r := range results{
		go func(r interface{}) error {
			defer wg.Done()
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
				rowsMutex.Lock()
				rowsToInsert = append(rowsToInsert, ticket)
				rowsMutex.Unlock()
				return nil
			}
			u.LogPrint(1, "Retrieving Routing Information")
			routingRows, err := b.GetRoutingRowsByProjectID(c.BqRoutingTable,row.Project_id)
			if err != nil {
				u.LogPrint(3,"Failed to get routing information")
				return err
			}
			ticket.TargetContact = routingRows[0].Target
			ticket.Assignee = routingRows[0].TicketSystemIdentifiers
			u.LogPrint(1,"Creating new Ticket")
			// I need a way to catch IF a ticket is already created
			ticketID, err := ticketService.CreateTicket(&ticket, row)
			if err != nil {
				return err
			}
			ticket.IssueKey = ticketID
			rowsMutex.Lock()
			rowsToInsert = append(rowsToInsert, ticket)
			rowsMutex.Unlock()
			return nil
		}(r)
	}
	wg.Wait()
	err = b.AppendTicketsToTable(c.BqTicketTable, rowsToInsert)
	if err != nil {
		u.LogPrint(3,err)
		return err
	}
	return err
}