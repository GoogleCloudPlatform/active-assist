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
	"log"
	"net/http"
	"os"
	b "ticketservice/internal/bigqueryfunctions"
	t "ticketservice/internal/ticketinterfaces"
	u "ticketservice/internal/utils"

	"github.com/codingconcepts/env"
	"github.com/labstack/echo/v4"
)

type config struct {
	BqDataset string `env:"BQ_DATASET" required:"true"`
	BqProject string `env:"BQ_PROJECT" required:"true"`
	BqRecommendationsTable string `env:"BQ_RECOMMENDATIONS_TABLE" default:"flattened_recommendations"`
	BqTicketTable	string `env:"BQ_TICKET_TABLE" default:"recommender_ticket_table"`
	BqRoutingTable	string `env:"BQ_ROUTING_TABLE" default:"recommender_routing_table"`
	TicketImpl	string `env:"TICKET_SERVICE_IMPL" default:"slackTicket"` //Needs to be the same name as the file without the extension
	TicketCostThreshold int `env:"TICKET_COST_THRESHOLD" default:"100"`
	AllowNullCost bool `env:"ALLOW_NULL_COST" default:"false"`
	ExcludeSubTypes string `env:"EXCLUDE_SUB_TYPES" default:"' '"` // Use commas to seperate
}

var c config
var ticketService t.BaseTicketService

// Init function for startup of application
func init() {
	// Print Startup so we know it's not lagging
	log.SetOutput(os.Stdout)
	u.LogPrint(1, "Ticket Service Starting")
	//Load env variables using "github.com/codingconcepts/env"
	if err := env.Set(&c); err != nil {
		u.LogPrint(4,err)
	}
	//initialize BigQuery
	b.InitBQ(c.BqDataset, c.BqProject)
	//Check For Access and Existence of BQ Table.
	u.LogPrint(1, "Creating Ticket Table")
	err := b.CreateOrUpdateTicketTable(c.BqTicketTable)
	if err != nil {
		log.Fatal(err)
	}
	u.LogPrint(1, "Creating Routing Table")
	err = b.CreateOrUpdateRoutingTable(c.BqRoutingTable)
	if err != nil {
		log.Fatal(err)
	}
	ticketService, err = t.InitTicketService(c.TicketImpl)
	if err != nil {
		u.LogPrint(4,"Failed to load ticket service plugin", err)
	}
}

func main() {

	e := echo.New()

	e.GET("/CreateTickets", func(c echo.Context) error {
		err := checkAndCreateNewTickets()
		if err != nil{
			u.LogPrint(3,"Error creating new ticket: %v",err)
			return err
		}
		return nil
	})

	// Create a new ticket.
	e.POST("/tickets", func(c echo.Context) error {
		var ticket t.Ticket
		if err := c.Bind(&ticket); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": err.Error(),
			})
		}
		issueKey, err := ticketService.CreateTicket(&ticket, t.RecommendationQueryResult{})
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": err.Error(),
			})
		}
		u.LogPrint(1,issueKey)

		return c.NoContent(http.StatusCreated)
	})

	// Close a ticket.
	e.PUT("/tickets/:issueKey/close", func(c echo.Context) error {
		// Extract issueKey
		var issueKey = c.Param("issueKey")

		// Check to make sure the ticket exists before continuing
		_, err := ticketService.GetTicket(issueKey)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				// Gonna need to think if this is ok to send back.
				"error": err.Error(),
			})
		}

		if err := ticketService.CloseTicket(issueKey); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": err.Error(),
			})
		}

		return c.NoContent(http.StatusNoContent)
	})

	// Handle webhook actions.
	e.POST("/webhooks/:action", func(c echo.Context) error {
		if err := ticketService.HandleWebhookAction(c); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": err.Error(),
			})
		}

		return c.NoContent(http.StatusOK)
	})

	// Start the server.
	if err := e.Start(":8080"); err != nil {
		log.Fatal(err)
	}
}
