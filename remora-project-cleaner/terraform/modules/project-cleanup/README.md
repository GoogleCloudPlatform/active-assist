# Google Cloud Project Cleanup Terraform Module

This module allows you to leverage recommendations made by the [Recommendations API](https://cloud.google.com/recommender/docs/reference/rest/v1/projects.locations.recommenders.recommendations) to alert project owners of unattended projects and subsequently delete them automatically. This is accomplished by using a service account to execute [Google Workflows](https://cloud.google.com/workflows) on a schedule to discover unattended projects in a given organization and send notification emails using [Sendgrid](https://sendgrid.com/) or open a ticket using [Jira](https://www.atlassian.com/software/jira).   

## Usage

See the [examples directory](../examples/) for additional examples of how to use.  The following will create the recommendations workflow using Sendgrid for notifications:
```hcl
module "recommendations_workflow_sendgrid" {
  source                 = "github.com/GoogleCloudPlatform/active-assist//remora-project-cleaner/terraform/modules/project-cleanup"
  project_id             = "my-project-id"
  organization_id        = "1234567890"
  cloudfunction_notifier = "sendgrid"
  sendgrid_config = {
    sendgrid_secret_name    = "my-sendgrid-key-secret"
    sendgrid_reply_to_email = "replyto@myorg.com"
    sendgrid_sender_email   = "project-cleanup@myorg.com"
  }
  time_zone     = "America/Los_Angeles"
  is_dry_run    = false
  schedule      = "0 9 * * 1"
  allow_metrics = true
}
```
## Prerequisites

The service account or user that will execute this module will need the following organization IAM roles for the organization whose project recommendations will be parsed:
* Organization Administrator

And the following project IAM roles for the project that will be used to host components:
* Cloud Functions Developer
* Cloud Functions Service Agent
* Cloud Scheduler Admin
* Project IAM Admin
* Pub/Sub Editor
* Secret Manager Admin
* Service Account Admin
* Service Account User
* Service Usage Admin
* Workflows Editor

Additionally, if using Sendgrid, a Google Secret Manager secret containing your Sendgrid API key will need to be created prior to running the module:
```sh
gcloud secrets create [KEY_NAME] --replication-policy="automatic" && gcloud secrets versions add [KEY_NAME] --data-file=[SENDGRID_API_KEY_PATH]
```
\[KEY_NAME\] will be provided as the argument `sendgrid_secret_name` in the `sendgrid_config` block.

Simililarly, if using Jira, a Google Secret Manager secret containing your Jira API key will need to be created prior to running the module:

```sh
gcloud secrets create [KEY_NAME] --replication-policy="automatic" && gcloud secrets versions add jira-key --data-file=[JIRA_API_KEY_PATH]
```
\[KEY_NAME\] will be provided as the argument `jira_secret_name` in the `jira_config` block. 


## Requirements

| Name | Version |
|------|---------|
| <a name="requirement_terraform"></a> [terraform](#requirement\_terraform) | >= 0.13 |
| <a name="requirement_google"></a> [google](#requirement\_google) | >= 3.53, < 5.0 |

## Providers

| Name | Version |
|------|---------|
| <a name="provider_archive"></a> [archive](#provider\_archive) | n/a |
| <a name="provider_google"></a> [google](#provider\_google) | >= 3.53, < 5.0 |
| <a name="provider_random"></a> [random](#provider\_random) | n/a |

## Modules

| Name | Source | Version |
|------|--------|---------|
| <a name="module_recs_workflow_custom_org_role"></a> [recs\_workflow\_custom\_org\_role](#module\_recs\_workflow\_custom\_org\_role) | terraform-google-modules/iam/google//modules/custom_role_iam | 7.4.0 |
| <a name="module_recs_workflow_custom_proj_role"></a> [recs\_workflow\_custom\_proj\_role](#module\_recs\_workflow\_custom\_proj\_role) | terraform-google-modules/iam/google//modules/custom_role_iam | 7.4.0 |

## Resources

| Name | Type |
|------|------|
| [google_cloud_scheduler_job.recommendations_workflow_run](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/cloud_scheduler_job) | resource |
| [google_cloudfunctions_function.create_jira_issue](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/cloudfunctions_function) | resource |
| [google_cloudfunctions_function.send_email](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/cloudfunctions_function) | resource |
| [google_project_service.remora_services](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/project_service) | resource |
| [google_pubsub_subscription.recommendations_workflow](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/pubsub_subscription) | resource |
| [google_pubsub_topic.recommendations_workflow](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/pubsub_topic) | resource |
| [google_service_account.remora_service_account](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/service_account) | resource |
| [google_storage_bucket.recommendation_workflow_functions](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/storage_bucket) | resource |
| [google_storage_bucket_object.recommendation_workflow_function_zip](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/storage_bucket_object) | resource |
| [google_workflows_workflow.recommendations_workflow_initial_setup](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/workflows_workflow) | resource |
| [google_workflows_workflow.recommendations_workflow_main](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/workflows_workflow) | resource |
| [google_workflows_workflow.recommendations_workflow_process_recommendations](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/workflows_workflow) | resource |
| [google_workflows_workflow.recommendations_workflow_summarize_and_notify](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/workflows_workflow) | resource |
| [random_id.role_id_suffix](https://registry.terraform.io/providers/hashicorp/random/latest/docs/resources/id) | resource |
| [archive_file.recommendation_workflow_function](https://registry.terraform.io/providers/hashicorp/archive/latest/docs/data-sources/file) | data source |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_allow_metrics"></a> [allow\_metrics](#input\_allow\_metrics) | If true, processing information will be stored as part of the recommendation's stateMetadata to be used to improve and measure the performance of Remora. The value of this field will not impact the functionality of Remora. | `bool` | n/a | yes |
| <a name="input_cloudfunction_notifier"></a> [cloudfunction\_notifier](#input\_cloudfunction\_notifier) | The desired method of notifying project owners eminent project deletion. Supported values are 'sendgrid' and 'jira' | `string` | n/a | yes |
| <a name="input_enable_apis"></a> [enable\_apis](#input\_enable\_apis) | Toggle to include required APIs. | `bool` | `false` | no |
| <a name="input_essential_contact_categories"></a> [essential\_contact\_categories](#input\_essential\_contact\_categories) | Categories for the Essential Contacts that should be notified during escalation. If none provided, the parent owner will used. | `list(string)` | `[]` | no |
| <a name="input_is_dry_run"></a> [is\_dry\_run](#input\_is\_dry\_run) | If true, will not delete project when it's the 3rd+ time seeing a recommendation and the TTL date has passed. If false, will delete the project when it's the 3rd+ time seeing a recommendation and the TTL date has passed. | `bool` | n/a | yes |
| <a name="input_jira_config"></a> [jira\_config](#input\_jira\_config) | Map of values required for Jira configuration: 'jira\_secret\_name', 'jira\_admin\_email', 'jira\_server\_address', 'jira\_project\_id', 'jira\_assignee\_account\_id' | `map(any)` | <pre>{<br>  "jira_admin_email": "",<br>  "jira_assignee_account_id": "",<br>  "jira_project_id": "",<br>  "jira_secret_name": "",<br>  "jira_server_address": ""<br>}</pre> | no |
| <a name="input_opt_out_project_numbers"></a> [opt\_out\_project\_numbers](#input\_opt\_out\_project\_numbers) | List of projects that won't be processed in the recommendations workflow. e.g. [111111111111, 222222222222] | `list(number)` | `[]` | no |
| <a name="input_organization_id"></a> [organization\_id](#input\_organization\_id) | The organization ID whose Unattended Project Recommendations will be used for the Recommendations Workflow. | `string` | n/a | yes |
| <a name="input_project_id"></a> [project\_id](#input\_project\_id) | The project ID where the Recommendations Workflow will be created. | `string` | n/a | yes |
| <a name="input_region"></a> [region](#input\_region) | Region to use for BigQuery, Cloud Workflows, functions, and scheduler. | `string` | `"us-central1"` | no |
| <a name="input_schedule"></a> [schedule](#input\_schedule) | Time interval defined using a unix-cron (see https://cloud.google.com/scheduler/docs/configuring/cron-job-schedules) format. This represents how frequently the recommendations workflow will run. | `string` | `"0 9 * * 1"` | no |
| <a name="input_sendgrid_config"></a> [sendgrid\_config](#input\_sendgrid\_config) | Map of values required for Sendgrid configuration: 'sendgrid\_secret\_name', 'sendgrid\_reply\_to\_email','sendgrid\_sender\_email' | `map(any)` | <pre>{<br>  "sendgrid_reply_to_email": "",<br>  "sendgrid_secret_name": "",<br>  "sendgrid_sender_email": ""<br>}</pre> | no |
| <a name="input_time_zone"></a> [time\_zone](#input\_time\_zone) | The time zone database name, which will be the time zone used for formatting and for the TTL date. A list of time zone database names can be found at https://en.wikipedia.org/wiki/List_of_tz_database_time_zones. e.g. America/Los\_Angeles | `string` | n/a | yes |
| <a name="input_ttl_days"></a> [ttl\_days](#input\_ttl\_days) | The number of days after which an unused project can be safely removed | `string` | `"30"` | no |

## Outputs

No outputs.