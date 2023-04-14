package main

import (
	"fmt"
	"os"
	bigqueryfunctions "ticketservice/internal/bigqueryfunctions"
)

//This probably should go in a utils file instead.
type EnvError struct {
    message string
}

func (e *EnvError) Error() string {
    return e.message
}

var envError = EnvError{"BIG_QUERY_PROJCET environment variable not set"}

func CheckAndCreateNewTickets() error {
	bigQueryProject := os.Getenv("BIG_QUERY_PROJECT")
	if bigQueryProject == "" {
		fmt.Println("BIG_QUERY_PROJCET environment variable not set")
		return &envError
	}
	// This is the function that should have parameters in it. 
	// Need to decide how to set those. 
	results, err := bigqueryfunctions.QueryBigQuery(
		bigQueryProject, 
		fmt.Sprintf("Select * from %s where something", "simple placeholder code"),
	)
	//Do something with the results
	for _, row := range results{
		// Create ticket here
		// This involves creating the ticket in ticketInterface
		// And then adding to BQ Table.
		fmt.Println(row[0])
	}
	return err
}