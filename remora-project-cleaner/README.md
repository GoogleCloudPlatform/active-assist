[![Linter](https://github.com/binamov/remora/actions/workflows/linter.yml/badge.svg?branch=main)](https://github.com/binamov/remora/actions/workflows/linter.yml)

# remora

_From Wikipedia:_

> The remora, sometimes called suckerfish, spend most of their lives clinging to a host animal such as a whale, turtle, shark or ray, removing ectoparasites and loose flakes of skin.

Remora automates the lifecycle management of unused projects in a cloud organization. Core capabilities include:

- continuous detection of unused projects
- notifications to the respective project Owners
- tracking of notification state
- escalation of notifications to folder/organization Admins
- removal of unused projects if notifications are not acted upon within a defined Time to Live (TTL)

## Disclaimer

This project is not an official Google project. It is not supported by
Google and Google specifically disclaims all warranties as to its quality,
merchantability, or fitness for a particular purpose.

## Usage

### General description

Remore uses a Service Account to run workflows, on a schedule (for example, every two weeks). Those workflows interact with the Recommendar API to discover unused projects in the organization. Workflows then identify the owners of those projects and send them an email.

Remora sends an email every time it runs, up to a maximum of three emails for any given project. 

Remora keeps track of emails for each project (using BigQuery). If a project continues to be unused after the first email to the owner(s), remora sends a copy of the next email to the folder or organization owner, whoever is the immediate parent of the project.

You get to set a time-to-live `TTL` which is the number of days after which an unused project can be safely removed. Remora labels every project with its impending deletion date.

*Remora removes projects when it runs and determines that: three sets of emails were sent, the TTL has expired, and the project is still unused.*

Remora can use sendgrid to send emails, it can also create Jira tickets. This is easy to modify and use a different provider.

### APIs and products used

#### Google APIs and products used
- [Recommender](https://cloud.google.com/recommender)
- [Scheduler](https://cloud.google.com/scheduler)
- [Workflows](https://cloud.google.com/workflows)
- [Functions](https://cloud.google.com/functions)
- [Pubsub](https://cloud.google.com/pubsub)
- [BigQuery](https://cloud.google.com/bigquery)
- [Secret Manager](https://cloud.google.com/secret-manager)
- [Cloud Resource Manager API](https://cloud.google.com/resource-manager/reference/rest)
- [Cloud Asset API](https://cloud.google.com/asset-inventory/docs/reference/rest)

The Google APIs and Products used follow the [Google Cloud Privacy Notice](https://cloud.google.com/terms/cloud-privacy-notice).

#### Non-Google APIs and products used
- [Sendgrid](https://sendgrid.com/)
- [Jira](https://www.atlassian.com/software/jira)


### IAM Permissions

The Service Account used by remora needs the following permissions:

- bigquery.datasets.create
- bigquery.datasets.get
- bigquery.jobs.create
- bigquery.tables.create
- bigquery.tables.get
- bigquery.tables.getData
- bigquery.tables.updateData
- cloudasset.assets.searchAllIamPolicies
- essentialcontacts.contacts.get
- essentialcontacts.contacts.list
- pubsub.topics.publish
- recommender.resourcemanagerProjectUtilizationInsights.get
- recommender.resourcemanagerProjectUtilizationInsights.list
- recommender.resourcemanagerProjectUtilizationRecommendations.get
- recommender.resourcemanagerProjectUtilizationRecommendations.list
- recommender.resourcemanagerProjectUtilizationRecommendations.update
- resourcemanager.projects.delete
- resourcemanager.projects.get
- resourcemanager.projects.getIamPolicy
- resourcemanager.projects.list
- resourcemanager.projects.update
- secretmanager.versions.access
- workflows.executions.create
- workflows.executions.get

To remove the IAM owner, the following permission is also needed:
- resourcemanager.projects.setIamPolicy

### Diagram
![diagram](../../assets/diagram.png)

### Getting started

You can follow the manual CLI step-by-step guide to configure the project and deploy remora assets. 
You can also follow the [Terraform step-by-step guide](./terraform/README.md) to configure the project and deploy remora assets.

#### CLI Step-by-step

1a.  If you are using Sendgrid, set up a Sendgrid account and generate an API key to be used to send emails.

1b.  If you are using Jira, generate a Jira API key for your Jira project.

2.  Set the following environment variables.

```
# Name of the service account to use for running the Recommendations Workflow.
# e.g. recs-workflow-sa
export SERVICE_ACCOUNT_NAME=<service-account-name>

# The organizationId whose Unattended Project Recommendations will be used for
# this Recommendations Workflow.
# e.g. 111111111111
export ORGANIZATION_ID=<org-id>

# Project to hold the Recommendations Workflow.
# e.g. test-project-111111
export PROJECT_ID=<project-id>

# Generates the email of the service account using the service account name
# and project id.
export SERVICE_ACCOUNT_EMAIL=${SERVICE_ACCOUNT_NAME}@${PROJECT_ID}.iam.gserviceaccount.com

# Region to use for the BigQuery dataset, BigQuery table, Cloud workflows, Cloud functions, and Cloud scheduler.
# e.g. us-east1
export REGION=<region>

# Time interval defined using a unix-cron (see https://cloud.google.com/scheduler/docs/configuring/cron-job-schedules)
# format. This represents how frequently the recommendations workflow will run.
# Please note that when setting this variable, you need to surround
# the frequency with single quotes, due to issues with the asterisk character.
# e.g. '0 9 * * 1'
export FREQUENCY='<frequency>'

# If using Sendgrid, path to the text file that contains the Sendgrid API key.
# e.g. path/to/text_file.txt
export SENDGRID_API_KEY_PATH=<sendgrid-api-key-path>

# If using Sendgrid, the sender email address.
# e.g. myEmail@company.com
export SENDGRID_SENDER_EMAIL=<sendgrid-send-email-address>

# If using Sendgrid, the email address that the reply message
# is sent to.
# e.g. replyToAddress@company.com
export SENDGRID_REPLY_TO_EMAIL=<sendgrid-reply-to-email-address>

# If using Jira, path to the text file that contains the Jira API key.
# e.g. path/to/text_file.txt
export JIRA_API_KEY_PATH=<jira-api-key-path>

# If using Jira, the email adress of the administrator. This should
# be the email address that was used to generate the API key. The associated
# user will become the reporter of the Jira issue.
# e.g. jiraAdmin@example.com
export JIRA_ADMINISTRATOR_EMAIL=<jira-administrator-email>

# If using Jira, the address of the Jira server.
# e.g. https://myJiraAddress.atlassian.net/
export JIRA_SERVER_ADDRESS=<jira-server-address>

# If using Jira, the id of the Jira project to create issues in.
# e.g. 10000
export JIRA_PROJECT_ID=<jira-project-id>

# If using Jira, the id of the Jira user who will be assigned issues.
# e.g. abcd1234efgh5678
export JIRA_ASSIGNEE_ACCOUNT_ID=<jira-assignee-account-id>

# How many days until a project can be deleted.
# e.g. 30
export NUM_DAYS_TTL=<num-days-ttl>

# The time zone database name, which will be the time zone used for formatting and 
# for the TTL date. A list of time zone database names can be found at 
# https://en.wikipedia.org/wiki/List_of_tz_database_time_zones.
# e.g. America/Los_Angeles
export TIME_ZONE_NAME=<time-zone-name>

# List of projects that won't be processed in the recommendations workflow.
# e.g. "[222222222222, 333333333333]"
export OPT_OUT_PROJECT_NUMBERS="[<opt-out-project-numbers>]"

# If true, will not delete a project when it's the 3rd+ time seeing a recommendation and the TTL date
# has passed. If false, will delete the project when it's the 3rd+ time seeing a recommendation and the 
# TTL date has passed. 
# e.g. true
export IS_DRY_RUN=<is-dry-run>

# During the second time a recommendation has been processed, we escalate to include the
# escalation contacts. This field is for what categories for the Essential Contacts we should
# notify during escalation. If we don't find any Essential Contacts for the categories, we will
# default to the parent owner. So, if you don't have essential contacts set up, you can use
# an empty list, and we'll just default to the parent owner.
# e.g. "[\\\"SECURITY\\\"]"
export ESSENTIAL_CONTACT_CATEGORIES="[<essential-contact-categories>]"

# If true, will store processing information as part of the recommendation's stateMetadata.
# If false, will not store processing information as part of the recommendation's stateMetadata.
# If stored, the processed information will be used to improve and measure the performance of Remora.
# Setting this field to true or false should not impact the functionality of Remora.
# e.g. true
export ALLOW_METRICS=<allow-metrics>
```

3.  (Optional) If you want to use a new project, create a new project and link billing.
```
# If creating under an organization:
gcloud projects create ${PROJECT_ID} --organization=${ORGANIZATION_ID}
# If creating under a folder:
gcloud projects create ${PROJECT_ID} --folder=<FOLDER_ID>
# Link billing:
gcloud beta billing projects link ${PROJECT_ID} --billing-account 0X0X0X-0X0X0X-0X0X0X
```

4.  Set gcloud to the project you want to use to host the recommendations workflow.
```
gcloud config set project ${PROJECT_ID}
```

5.  Enable relevant APIs.
```
gcloud services enable recommender.googleapis.com
gcloud services enable workflows.googleapis.com
gcloud services enable cloudfunctions.googleapis.com
gcloud services enable pubsub.googleapis.com
gcloud services enable bigquery.googleapis.com
gcloud services enable secretmanager.googleapis.com
gcloud services enable cloudresourcemanager.googleapis.com
gcloud services enable cloudasset.googleapis.com
gcloud services enable cloudbuild.googleapis.com
gcloud services enable cloudscheduler.googleapis.com
gcloud services enable essentialcontacts.googleapis.com
```

6.  Create a Service Account with permissions to run the recommendations workflow.
```
gcloud iam service-accounts create ${SERVICE_ACCOUNT_NAME}

gcloud iam roles create recs_workflow_org_sa_permissions --organization=${ORGANIZATION_ID} \
    --file=permissions/recs_workflow_org_sa_permissions.yaml
gcloud organizations add-iam-policy-binding ${ORGANIZATION_ID} \
    --member=serviceAccount:${SERVICE_ACCOUNT_EMAIL} \
    --role=organizations/${ORGANIZATION_ID}/roles/recs_workflow_org_sa_permissions 
    
gcloud iam roles create recs_workflow_proj_sa_permissions --project=${PROJECT_ID} \
    --file=permissions/recs_workflow_proj_sa_permissions.yaml
gcloud projects add-iam-policy-binding ${PROJECT_ID} \
    --member=serviceAccount:${SERVICE_ACCOUNT_EMAIL} \
    --role=projects/${PROJECT_ID}/roles/recs_workflow_proj_sa_permissions 
```

7.  Deploy Cloud workflows.
```
gcloud beta workflows deploy recommendations_workflow_main --source=terraform/modules/project-cleanup/workflows/recommendations_workflow_main.yaml \
    --service-account=${SERVICE_ACCOUNT_EMAIL} --location=${REGION}
gcloud beta workflows deploy recommendations_workflow_initial_setup \
    --source=terraform/modules/project-cleanup/workflows/recommendations_workflow_initial_setup.yaml \
    --service-account=${SERVICE_ACCOUNT_EMAIL} --location=${REGION}
gcloud beta workflows deploy recommendations_workflow_process_recommendations \
    --source=terraform/modules/project-cleanup/workflows/recommendations_workflow_process_recommendations.yaml \
    --service-account=${SERVICE_ACCOUNT_EMAIL} --location=${REGION}
gcloud beta workflows deploy recommendations_workflow_summarize_and_notify \
    --source=terraform/modules/project-cleanup/workflows/recommendations_workflow_summarize_and_notify.yaml \
    --service-account=${SERVICE_ACCOUNT_EMAIL} --location=${REGION}
```

8.  Create pub-sub to connect recommendations_workflow_summarize_and_notify to
    main.py.
```
gcloud pubsub topics create recommendations-workflow-topic
gcloud pubsub subscriptions create recommendations-workflow--subscription \
    --topic recommendations-workflow-topic
```

9a.  If using Sendgrid, set up secret for accessing the Sendgrid API key. 
```
gcloud secrets create sendgrid-key \
    --replication-policy="automatic"
gcloud secrets versions add sendgrid-key --data-file=${SENDGRID_API_KEY_PATH}
```

9b. If using Jira, set up secret for accessing the Jira API key.
```
gcloud secrets create jira-key \
    --replication-policy="automatic"
gcloud secrets versions add jira-key --data-file=${JIRA_API_KEY_PATH}
```

10a.  If using Sendgrid, deploy the Cloud function to send emails via Sendgrid.
```
gcloud beta functions deploy send_email --runtime=python38 --source=terraform/modules/project-cleanup/functions/sendgrid/ \
    --trigger-topic=recommendations-workflow-topic \
    --set-env-vars SENDER_EMAIL=${SENDGRID_SENDER_EMAIL},REPLY_TO_EMAIL=${SENDGRID_REPLY_TO_EMAIL} \
    --set-secrets='sendgrid-key=sendgrid-key:1' --service-account=${SERVICE_ACCOUNT_EMAIL}
```

10b.  If using Jira, deploy the Cloud function to create Jira issues.
```
gcloud beta functions deploy create_jira_issue --runtime=python38 --source=terraform/modules/project-cleanup/functions/jira/ \
    --trigger-topic=recommendations-workflow-topic \
    --set-env-vars JIRA_ADMINISTRATOR_EMAIL=${JIRA_ADMINISTRATOR_EMAIL},JIRA_SERVER_ADDRESS=${JIRA_SERVER_ADDRESS},JIRA_PROJECT_ID=${JIRA_PROJECT_ID},JIRA_ASSIGNEE_ACCOUNT_ID=${JIRA_ASSIGNEE_ACCOUNT_ID} \
    --set-secrets='jira-key=jira-key:1' --service-account=${SERVICE_ACCOUNT_EMAIL}
```

11.  Create an App Engine app (only run this if one doesn't already exist).
```
# There is an issue for us-central1 and europe-west1, --region only supports us-central
# and europe-west, so if ${REGION} is us-central1 or europe-west1, you'll need to use 
# us-central or europe-west instead.
gcloud app create --region=${REGION}
```

12.  Schedule running the Recommendations Workflow.
```
gcloud scheduler jobs create http scheduled_recommendations_workflow_run \
--schedule="${FREQUENCY}" \
--uri=https://workflowexecutions.googleapis.com/v1/projects/${PROJECT_ID}/locations/${REGION}/workflows/recommendations_workflow_main/executions \
--message-body="{\"argument\": \"{\\\"numDaysTTL\\\": ${NUM_DAYS_TTL}, \\\"region\\\": \\\"${REGION}\\\", \\\"timeZoneName\\\": \\\"${TIME_ZONE_NAME}\\\", \\\"organizationId\\\": ${ORGANIZATION_ID}, \\\"optOutProjectNumbers\\\": ${OPT_OUT_PROJECT_NUMBERS}, \\\"essentialContactCategories\\\": ${ESSENTIAL_CONTACT_CATEGORIES}, \\\"isDryRun\\\": ${IS_DRY_RUN}, \\\"allowMetrics\\\": ${ALLOW_METRICS}}\"}" \
--time-zone=${TIME_ZONE_NAME} \
--location=${REGION} \
--oauth-service-account-email=${SERVICE_ACCOUNT_EMAIL}
```


## Contributing

See [`CONTRIBUTING.md`](CONTRIBUTING.md) for details.

## License

Apache 2.0; see [`LICENSE`](LICENSE) for details.

