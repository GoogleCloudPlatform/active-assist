# Ticket Service

This is a Go based microservice designed to handle ticketing logic. It includes features like creating, closing tickets and managing webhooks for ticket operations. The service also interacts with Google's BigQuery for data handling. 

## Getting Started

These instructions will get you a copy of the project up and running on your local machine for development and testing purposes.

### Prerequisites

Ensure that you have the latest version of Go installed on your machine. The code is tested with Go version 1.19. However, it should work with newer versions as well.

**Note:**  The current configuration here assumes you are using Application Default Credentials for BQ Access.

You also need to have Recommender API exports and Asset Inventory exports configured to send to BigQuery. [Please see our Workflow for automating this process](org-level-recommendations-hub/workflows)

### Installing

- Clone the repository to your local machine
```
git clone https://github.com/Crash-GHaun/active-assist.git
```

- Navigate into the cloned repository
```
cd org-level-recommendations-hub/ticketservice
```

- Download Dependencies
```
go mod download
```

-- Compile the plugins
```
go run compilePlugins.go
```

- To start the service locally, run the following command
```
go run main.go ticketService.go
```

## Configuration

Configuration is handled through environment variables. The list of required and optional environment variables are:

- BQ_DATASET **(required)**
  - BigQuery Dataset that contains exported recommendations
- BQ_PROJECT **(required)**
  - BigQuery Project your dataset is in.
- BQ_RECOMMENDATIONS_TABLE (optional, defaults to "flattened_recommendations")
  - Table/View name of your [Flattened Recommendations](org-level-recommendations-hub/flatten-table-bigquery.sql)
- BQ_TICKET_TABLE (optional, defaults to "recommender_ticket_table")
  - The name of the table you want to use for storing ticket data
- BQ_ROUTING_TABLE (optional, defaults to "recommender_routing_table")
  - The name of the table that stores project to target and system identifiers. See [Routing Table](#routing-table) for more information.
- TICKET_SERVICE_IMPL (optional, defaults to "slackTicket")
  - The Ticket Service Implementation you want to use. I.E (slackTicket). This should match the name of the plugin without the .so extension.
- TICKET_COST_THRESHOLD (optional, defaults to 100)
  - Limits the creation of tickets to a certain monetary threshold. 
- TICKET_LIMIT (optional, defaults to 5)
  - You can limit the amount of tickets created per call to reduce spam
- ALLOW_NULL_COST (optional, defaults to "false")
  - This allows you to create tickets for recommendations that **do not** have costs associated with them.
- EXCLUDE_SUB_TYPES (optional, defaults to ' ')
  - A Comma seperated list that allows you to filter the types of recommendations that recieve tickets.

Please note that the environment variables needs to be set before starting the service.

## Endpoints

- `GET /CreateTickets`: Checks for new tickets, and Updates stale tickets.
- `POST /tickets`: Creates a new ticket.
- `PUT /tickets/:issueKey/close`: Closes an existing ticket.
- `POST /webhooks`: Handles webhook actions based on your ticket service.

Sure, here's a Deployment section you can add to your `README.md`:


## Deployment

This service is deployed using Google Cloud Build and Docker. 

### Docker

A Dockerfile is included in the repository. The Dockerfile uses a two-stage build process. In the first stage, it compiles the Go code to create an executable.It also compiles the plugins included in this repo.

In the second stage, it copies the compiled binary into a new Docker image.


### Google Cloud Build

Cloud Build is configured to build the Dockerfile, push the image to Google Container Registry, and then deploy the image to Google Cloud Run with some environment variables set. To use this build file you must set the following substitutions.

- _SERVICE
  - Name of the Cloud Run Service
- _REGION
- _BQ_DATASET
- _LOG_LEVEL
- _SLACK_CHANNEL_AS_TICKET
- _TICKET_COST_THRESHOLD

If you are using the Slack integration you will need the following secrets configured.

- SLACK_SIGNING_SECRET
- SLACK_API_TOKEN

## Routing Table

The Ticket Service relies on a BigQuery table for routing tickets to the appropriate person or team. This table contains the following schema:

```go
bigquery.Schema{
    {Name: "Target", Type: bigquery.StringFieldType, Required: true},
    {Name: "ProjectID", Type: bigquery.StringFieldType},
    {Name: "TicketSystemIdentifiers", Type: bigquery.StringFieldType, Repeated: true},
}
```

### Ticket Routing

As of the current version, all routing of recommendations is done based on the `ProjectID`. Each recommendation gets mapped to a `ProjectID`, which then provides the necessary routing information for ticket creation.

### Target Field

The `Target` field is determined by the desired location or component where the ticket will be created. This is based on the specific ticket implementation in use.

For example, if you are using Slack (with the `SLACK_CHANNEL_AS_TICKET` environment variable set to `false`), the `Target` would be the Slack channel name where a thread should be initiated.

### TicketSystemIdentifiers Field

The `TicketSystemIdentifiers` is a repeated string field that directly corresponds to the "Assignees" in the ticketing system. 

For instance, in Slack, identifiers are not usernames or emails, but unique strings like `U03CS3FK54Z`. Therefore, this field should be configured based on the specifics of your ticketing system.

## License

This project is licensed under the Apache License - see the [LICENSE.md](LICENSE.md) file for details
