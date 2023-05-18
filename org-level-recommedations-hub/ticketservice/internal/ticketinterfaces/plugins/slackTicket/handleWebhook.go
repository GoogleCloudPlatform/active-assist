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
// limitations under the License.

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/slack-go/slack/slackevents"

	u "ticketservice/internal/utils"
)

func sendFastResponseAndProcess(body []byte, c echo.Context) error {
	// Get the URL from the request
	url := fmt.Sprintf("https://%s%s", c.Request().Host, c.Request().URL.Path)
	fmt.Printf("Resend to: %s\n", url)

	// Create a new request with the received URL
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}

	// Forward the request with an additional header
	req.Header.Set("X-PROCESS-SLACK", "true")

	// Add the X-Slack-Signature header
	signature := c.Request().Header.Get("X-Slack-Signature")
	timestamp := c.Request().Header.Get("X-Slack-Request-Timestamp")
	req.Header.Set("X-Slack-Signature", signature)
	req.Header.Set("X-Slack-Request-Timestamp", timestamp)

	// Create a custom RoundTripper to ignore the response body
	transport := &http.Transport{
		DisableKeepAlives: true,
	}
	client := &http.Client{Transport: transport}
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	go func() {
	// Perform the request, waiting for the request to finish sending
		_, err = client.Do(req)
		if err != nil {
			fmt.Println("Error:", err)
		}
	}()
	// Small Delay to ensure the request is completely sent.
	time.Sleep(time.Second)
	c.NoContent(http.StatusAccepted)
	return nil
}

func (s *SlackTicketService) HandleWebhookAction(c echo.Context) error {
	// So the problem with Slack is that they expect a response within 3 seconds
	// https://api.slack.com/apis/connections/events-api#responding
	// So we actually need to send a response quickly and process in the background.
	// BQ as our main datasource means it's impossible to respond within 3 seconds.

	// We MUST respond within 3 seconds or a retry happens, however we could solve
	// The retry issue with some form of caching, however that doesn't completely solve
	// the problem, because Slack will disable events if the majority of events are over 3s

	// Here is where it get's interesting, we want to support serverless deployment
	// Serverless services generally shut down processing once a response has been sent.
	// There are many ways to solve this problem, but in most cases it uses another outside service
	// Ideally this should be able to run on any platform, regardless of deployment.

	// I'm including it only in the Slack Plugin, because this may not be a problem for other plugins
	// Happy to move it if we need.

	// So to get around this, and NOT use other services (PubSub, Cloud Tasks) we will call
	// the Webhook again, but with a header that allows the service to actually process.

	//Verifying the request is ALWAYS first. 
	defer c.Request().Body.Close()
	body, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		return err
	}
	if !verifyRequestSignature(c.Request().Header, body) {
		return fmt.Errorf("Failed to Verify Request Signature")
	}

	    // Parse the event payload
    u.LogPrint(1, "Body: %v", string(body))
    var event slackevents.EventsAPICallbackEvent
    err = json.Unmarshal(body, &event)
    if err != nil {
        return err
    }
    switch event.Type {
    case slackevents.URLVerification:
        var r *slackevents.ChallengeResponse
        err := json.Unmarshal(body, &r)
        if err != nil {
            return err
        }
        return c.JSON(http.StatusOK, r)

    case slackevents.CallbackEvent:
		// Check if the X-PROCESS-SLACK header is set to true
		processSlack := c.Request().Header.Get("X-PROCESS-SLACK") == "true"
		if !processSlack {
			u.LogPrint(1, "Recieved first webhook, will process in background")
			sendFastResponseAndProcess(body, c)
			return c.NoContent(http.StatusOK)
		}
		var eventType *slackevents.EventsAPIInnerEvent
		err = json.Unmarshal([]byte(*event.InnerEvent), &eventType)
		if err != nil {
            return err
        }
        if slackevents.EventsAPIType(eventType.Type) == slackevents.Message {
            // Unmarshal the inner event into a MessageEvent
            var messageEvent *slackevents.MessageEvent
            err = json.Unmarshal(*event.InnerEvent, &messageEvent)
            if err != nil {
                return err
            }
            // Now you have access to the message event data
            u.LogPrint(1, "Received message event: %v", messageEvent)
			messageSplitBySpaces := strings.Split(messageEvent.Text, " ")
			if len(messageSplitBySpaces) < 1 {
				u.LogPrint(1, "Message did not have any length")
				return nil
			}
			// Check if it's a command
			command := strings.ToLower(messageSplitBySpaces[0])
			if !regexp.MustCompile(`^!`).MatchString(command) {
				u.LogPrint(1, "Not a command: %v", command)
				return nil
			}
			// Now let's fire up the correct command
			if function, ok := functionMap[command]; ok {
				// Call the function with the event and arguments
				err := function(s, messageEvent, messageSplitBySpaces)
				if err != nil {
					u.LogPrint(3, "Something went wrong with function: %v", command)
				}
				u.LogPrint(2, "Completed Function %v", command)
			} else {
				// Command not found
				u.LogPrint(1, "Command %v not found", command)
				return nil
			}
			return nil

        }

    default:
        return c.String(http.StatusInternalServerError, fmt.Sprintf("Unexpected event type: %s", event.Type))
    }
    return nil
}