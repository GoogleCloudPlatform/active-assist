package ticketService

import (
	"log"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()

	ticketService := &baseTicketService{}

	// Create a new ticket.
	e.POST("/tickets", func(c echo.Context) error {
		var ticket Ticket
		if err := c.Bind(&ticket); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": err.Error(),
			})
		}

		if err := ticketService.CreateTicket(ticket); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": err.Error(),
			})
		}

		return c.NoContent(http.StatusCreated)
	})

	// Close a ticket.
	e.PUT("/tickets/:id/close", func(c echo.Context) error {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": err.Error(),
			})
		}

		if err := ticketService.CloseTicket(id); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": err.Error(),
			})
		}

		return c.NoContent(http.StatusNoContent)
	})

	// Search for tickets.
	e.GET("/tickets/search", func(c echo.Context) error {
		query := c.QueryParam("q")

		tickets, err := ticketService.SearchTickets(query)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": err.Error(),
			})
		}

		return c.JSON(http.StatusOK, tickets)
	})

	// Handle webhook actions.
	e.POST("/webhooks/:action", func(c echo.Context) error {
		action := c.Param("action")

		if err := ticketService.HandleWebhookAction(action, nil); err != nil {
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
