# README for Slack Ticket Service

## Overview
This project contains a `SlackTicketService` written in Go that communicates with Slack using the Slack API to create and manage a ticketing system.

## Features
1. Authentication with the Slack API.
2. Capability to create and manage tickets either using channels or threads on Slack based on user preference.

## Requirements
1. Go (The Go Programming Language) installed on your machine.
2. Slack API token.
3. Slack signing secret.

## Environment Variables
This service uses the following environment variables:

1. `SLACK_API_TOKEN`: The API token for your Slack App. This is mandatory for the service to interact with Slack's API.
2. `SLACK_SIGNING_SECRET`: The signing secret for your Slack App. This is also mandatory for the service.
3. `SLACK_CHANNEL_AS_TICKET`: This is an optional variable. When set to true, the service will use channels as tickets. When set to false, it will use threads as tickets. If this environment variable is not set, it defaults to true.


## Creating a Slack App
Before you start, you'll need to create a Slack App and give it the required permissions:

1. Visit the [Slack API website](https://api.slack.com/apps?new_app=1) and click the 'Create New App' button.

2. Choose 'From scratch' and fill in the App Name and Development Slack Workspace fields, then click 'Create App'.

3. In the 'Basic Information' page, scroll down to 'App Credentials' and note down the 'Signing Secret'. Set this as your environment variable `SLACK_SIGNING_SECRET`.

4. Navigate to 'OAuth & Permissions' on the left-hand menu and scroll down to the 'Scopes' section. Here you can add the necessary scopes (permissions) to your app. Add the following scopes under 'Bot Token Scopes':
   - `channels:history`
   - `channels:join`
   - `channels:manage`
   - `channels:read`
   - `channels:write`
   
   Note: The app may require additional permissions depending on further requirements.

5. Navigate to 'Event Subscriptions' on the left-hand menu. Set it to 'On'. In the 'Request URL' field, you'll need to provide the public URL of the server where this service is running. Slack will send a verification request to this URL. If the service is running locally, consider using a service like ngrok to expose your local server.

6. Under 'Subscribe to Bot Events', click 'Add Bot User Event' and add the events you want your bot to listen to.

7. Go back to 'OAuth & Permissions', click 'Install App to Workspace'. Authorize the app in your workspace, after which you'll be provided with a 'Bot User OAuth Token'. Set this as your environment variable `SLACK_API_TOKEN`.
