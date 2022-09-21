#!/usr/bin/python
#
# Copyright 2022 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
"""
Generates Jira issues.
"""

import base64
import json
import os
from jira import JIRA

def create_jira_issue(event, context):
    """Generates Jira issues regarding unattended GCP projects.
     This Cloud function expects that Secret Manager is set up for
     the jira-key key, and that the jira-key key is being exposed as an
     environmental variable. This Cloud function also expects
     JIRA_ADMINISTRATOR_EMAIL, JIRA_SERVER_ADDRESS, JIRA_PROJECT_ID, and
     JIRA_ASSIGNEE_ACCOUNT_ID are populated environment variables.
    Args:
         event (dict):  The dictionary with data specific to this type of event.
         context (google.cloud.functions.Context): Metadata of triggering event.
    Returns:
        None. The output is written to Cloud Logging.
    """
    # Secret versions may have a new line character added to them,
    # so we remove the new line character at the end when we get
    # jira-key.
    jira_api_key = os.environ.get('jira-key',
    	'environment variable jira-key is not set.').rstrip('\n')
    pubsub_message = base64.b64decode(event['data']).decode('utf-8')
    message_json = json.loads(pubsub_message)

    description = ("You have Google Cloud project(s) that are not being used. "
        + "We recommend you remove those project(s). "
        + "Please review if the project(s) are still needed. "
        + "This is notice #" + message_json["numPasses"] + ". ")

    description += ("\n||Project Id||Usage Details||"
        + "TTL Date||Deleted||\n")
    for recommendation in message_json['recommendations']:
        description += ("|" + recommendation["projectId"]
            + "|[Review|"
            + "https://console.cloud.google.com/home/recommendations/view-link/"
            + "projects/" + recommendation["projectNumber"]
            + "/locations/global/recommenders/google.resourcemanager."
            + "projectUtilization.Recommender/recommendations/" + recommendation['recommendationId']
            + ";source=webSubtask?project=" + recommendation["projectId"]
            + "&e=ViewLinkLaunch::ViewLinkEnabled]")
        if message_json["isDryRun"] == "true":
            description += ("|n/a|n/a|\n")
        else:
            description += ("|" + recommendation["ttlFormattedTimestamp"]
                + "|" + recommendation["isDeleted"] + "|\n")
    if message_json["isDryRun"] == "true":
        description += ("Projects will be deleted after their TTL date.\n")
    administrator_email = os.environ.get('JIRA_ADMINISTRATOR_EMAIL',
    	'environment variable JIRA_ADMINISTRATOR_EMAIL is not set.')
    server_address = os.environ.get('JIRA_SERVER_ADDRESS',
    	'environment variable JIRA_SERVER_ADDRESS is not set.')
    jira = JIRA(basic_auth=(administrator_email, jira_api_key),
    	options={"server": server_address})

    project_id = os.environ.get('JIRA_PROJECT_ID',
    	'environment variable JIRA_PROJECT_ID is not set.')
    assignee_account_id = os.environ.get('JIRA_ASSIGNEE_ACCOUNT_ID',
    	'environment variable JIRA_ASSIGNEE_ACCOUNT_ID is not set.')
    # Data for creating a Jira issue.
    issue_data = {
        "project": {"id": project_id},
        "summary": "Unattended GCP Projects Were Detected",
        "description": description,
        "issuetype": {'name': 'Task'},
        "assignee": {
            "accountId": assignee_account_id
        }
    }

    try:
        # Create issue in Jira.
        jira.create_issue(issue_data)
    # pylint: disable=W0703
    except Exception as jira_exception:
        print(jira_exception)
        print(context)
