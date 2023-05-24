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

package ticketinterfaces

import (
	"time"
	"plugin"
	"log"
	"github.com/labstack/echo/v4"
)

// Your plugin needs to have the method CreateService that returns your BaseTicketService interface implementation

// TicketService is an interface for managing tickets.
type BaseTicketService interface {
	Init() error
	// I might want to update this to not return anything, except err. Because we are modifying the 
	// original variable anyways. 
	CreateTicket(ticket *Ticket, row RecommendationQueryResult) (string, error)
	UpdateTicket(ticket *Ticket, row RecommendationQueryResult) error
	CloseTicket(issueKey string) error
	GetTicket(issueKey string) (Ticket, error)
	HandleWebhookAction(echo.Context) error
}

// Ticket represents a support ticket.
type Ticket struct {
	IssueKey        string    `json:"issueKey"`
	TargetContact	string	  `json:"targetContact"`
	CreationDate    time.Time `json:"creationDate"`
	Status          string    `json:"status"`
	TargetResource  string    `json:"targetResource"`
	RecommenderID  string  `json:"recommenderIds"`
	LastUpdateDate time.Time `json:"lastUpdateDate"`
	LastPingDate    time.Time `json:"lastPingDate"`
	SnoozeDate      time.Time `json:"snoozeDate"`
	Subject         string    `json:"subject"`
	Assignee        []string    `json:"assignee"`
}


type RecommendationQueryResult struct {
	Project_name	string	
	Project_id	string
	Recommender_name	string
	Location	string
	Recommender_subtype	string
	Impact_cost_unit	int
	Impact_currency_code	string
	TargetResource	string
	Description	string
	Ticket	Ticket
}

func InitTicketService(implName string) (BaseTicketService, error) {

	// Load the plugin based on the name
	pluginPath := "./plugins/" + implName + ".so"
	p, err := plugin.Open(pluginPath)
	if err != nil {
		u.LogPrint(1, "Plugin name: %v", implName)
		u.LogPrint(4, "Failed to open plugin: %v", err)
	}

	// Look up the "NewTicketService" symbol in the plugin
	newTicketServiceSymbol, err := p.Lookup("CreateService")
	if err != nil {
		u.LogPrint(4, "Failed to lookup symbol: %v", err)
	}

	// Create an instance of the ticket service implementation
	implValue := newTicketServiceSymbol.(func() BaseTicketService)()

	// Initialize the ticket service implementation
	if err := implValue.Init(); err != nil {
		return nil, err
	}

	return implValue, nil
}

// %[1] is the recommender export table
// %[2] is the ticket table
// %[3] is the Cost Threshold
// %[4] is an additional string added to allow null values
// TODO: (GHAUN) reduce the number of returned fields
var CheckQueryTpl = `SELECT f.* EXCEPT(
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