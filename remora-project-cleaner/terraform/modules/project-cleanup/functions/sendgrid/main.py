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
Sends emails via Sendgrid.
"""

import base64
import json
import os
from sendgrid import SendGridAPIClient
from sendgrid.helpers.mail import Mail, Cc

def send_email(event, context):
    """Sends emails regarding unattended GCP projects to the project owner
     (and may CC the escalation contacts if the recommendations have been processed
     2+ times). This Cloud function expects that Secret Manager is set up for
     the Sendgrid key, and that the Sendgrid key is being exposed as an
     environmental variable. This Cloud function also expects
     SENDER_EMAIL and REPLY_TO_EMAIL are populated environment variables.
    Args:
         event (dict):  The dictionary with data specific to this type of event.
         context (google.cloud.functions.Context): Metadata of triggering event.
    Returns:
        None. The output is written to Cloud Logging.
    """
    message_json = json.loads(base64.b64decode(event['data']).decode('utf-8'))

    # Convert string version of the lists to actual lists.
    escalation_contacts = json.loads(message_json['escalationContacts'])
    project_contacts = json.loads(message_json["projectContacts"])

    # We need to remove any project contacts from the escalation
    # contacts list, because you're unable to send an email to someone
    # who is both a recipient and CC.
    escalation_contacts = list(set(escalation_contacts)-set(project_contacts))

    # Generate the table containing recommendations and keep track to see
    # if any projects were deleted.
    recommendations_string = ("<table><tr><td>Project Id</td><td>Usage Details</td>")
    if message_json["isDryRun"] == "false":
        recommendations_string += "<td>TTL Date</td><td>Was Deleted</td>"
    recommendations_string += "</tr>"
    any_projects_deleted = False

    for recommendation in message_json['recommendations']:
        recommendation_link = ("https://console.cloud.google.com/home/recommendations/"
            + "view-link/projects/" + recommendation["projectNumber"]
            + "/locations/global/recommenders/google.resourcemanager."
            + "projectUtilization.Recommender/recommendations/" + recommendation['recommendationId']
            + ";source=webSubtask?project=" + recommendation["projectId"]
            + "&e=ViewLinkLaunch::ViewLinkEnabled")
        recommendations_string += ("<tr><td>" + recommendation["projectId"]
       	    + "</td><td><a href=\""
       	    + recommendation_link + "\">Review</a></td>")
        if message_json["isDryRun"] == "false":
            recommendations_string += ("<td>" + recommendation["ttlFormattedTimestamp"]  +
                "</td><td>" + recommendation["isDeleted"] + "</td>")
        recommendations_string += "</tr>"
        if recommendation["isDeleted"] == "true":
            any_projects_deleted = True

    recommendations_string += ("</table>")

    # Generate the contents of the email.
    email = "You are an Owner on Google Cloud projects that are not being used. "

    if any_projects_deleted:
        email += ('Some of these projects have been deleted due to no action taken '
                + 'during the previous notices. <br/><br/>If you would like to recover '
                + 'any of the projects deleted, you can try restoring them with '
                + ' the steps listed <a href="https://cloud.google.com/resource-manager/'
                + 'docs/creating-managing-projects#restoring_a_project">here</a>.<br/><br/>')
    else:
        email += ('We recommend you remove those projects. '
            + 'Please review if the projects are still needed.<br/><br/>'
            + 'This is notice #' + message_json["numPasses"] + '.<br/><br/>')
        if message_json["isDryRun"] == "false":
            email += ('If projects are not deleted or recommendations are not dismissed, '
                    + 'projects will be deleted on or after their TTL date.<br/><br/>')

    if message_json['numPasses'] != '1':
        email += ((', '.join(escalation_contacts)) + " have been CCed, since they are"
                + " listed as the Essential Contacts or are the parent resource owner.<br/><br/>")


    email += recommendations_string

    # If there are no project contacts, default to the escalation contacts.
    if not project_contacts:
        project_contacts = escalation_contacts

    message = Mail(
        from_email=os.environ.get('SENDER_EMAIL',
    		'environment variable SENDER_EMAIL is not set.'),
        to_emails=project_contacts,
        subject='Unattended GCP Projects Were Detected',
        html_content=email)

    message.reply_to = os.environ.get('REPLY_TO_EMAIL',
    	'environment variable REPLY_TO_EMAIL is not set.')

    # Add escalation contacts to CC.
    if message_json['numPasses'] != '1':
        for escalation_contact in escalation_contacts:
            message.add_cc(Cc(escalation_contact, escalation_contact))

    try:
        # Secret versions may have a new line character added to them,
        # so we remove the new line character at the end when we get
        # the sendgrid key.
        sg_client = SendGridAPIClient(os.environ.get('sendgrid-key',
    	    'environment variable sendgrid-key is not set.').rstrip('\n'))
        response = sg_client.send(message)
        print(response.status_code)
        print(response.body)
        print(response.headers)
    # pylint: disable=W0703
    except Exception as sendgrid_exception:
        print(sendgrid_exception)
        print(context)
