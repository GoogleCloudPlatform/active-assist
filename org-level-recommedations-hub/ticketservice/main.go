package main

import (
	"fmt"
	"log"
	"net/http"
	t "ticketservice/internal/ticketinterfaces"

	"github.com/labstack/echo/v4"
)

// Curious if I should make a struct here
// define ticket interface and other stuff.

func main() {
	e := echo.New()

	// TODO(GHAUN): Make this variable depending on what plugin should be used.
	ticketService := &t.SlackTicketService{}

	// Create a new ticket.
	e.POST("/tickets", func(c echo.Context) error {
		var ticket t.Ticket
		if err := c.Bind(&ticket); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": err.Error(),
			})
		}
		issueKey, err := ticketService.CreateTicket(ticket)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": err.Error(),
			})
		}
		fmt.Println(issueKey)

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
