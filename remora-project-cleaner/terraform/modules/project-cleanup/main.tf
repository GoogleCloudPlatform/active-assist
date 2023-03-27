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

locals {
  sendgrid_cloudfunctions_environment_variables = {
    "REPLY_TO_EMAIL" = var.sendgrid_config.sendgrid_reply_to_email
    "SENDER_EMAIL"   = var.sendgrid_config.sendgrid_sender_email
  }
  jira_cloudfunctions_environment_variables = {
    "JIRA_ADMINISTRATOR_EMAIL" = var.jira_config.jira_admin_email
    "JIRA_SERVER_ADDRESS"      = var.jira_config.jira_server_address
    "JIRA_PROJECT_ID"          = var.jira_config.jira_project_id
    "JIRA_ASSIGNEE_ACCOUNT_ID" = var.jira_config.jira_assignee_account_id
  }
  service_account_id = "remora-service"
}

# Service Account and IAM

resource "google_service_account" "remora_service_account" {
  account_id   = local.service_account_id
  display_name = "Remora service account"
  depends_on = [
    google_project_service.remora_services
  ]
}

resource "random_id" "role_id_suffix" {
  byte_length = 2
}

module "recs_workflow_custom_proj_role" {
  source       = "terraform-google-modules/iam/google//modules/custom_role_iam"
  version      = "7.4.0"
  target_level = "project"
  stage        = "ALPHA"
  target_id    = var.project_id
  role_id      = format("recs_workflow_proj_sa_permissions_%s", random_id.role_id_suffix.hex)
  title        = "recs_workflow_proj_sa_permissions"
  description  = "Project level IAM permissions for the Service Account to run the Recommendations workflow."
  permissions = [
    "bigquery.datasets.create",
    "bigquery.datasets.get",
    "bigquery.jobs.create",
    "bigquery.tables.create",
    "bigquery.tables.get",
    "bigquery.tables.getData",
    "bigquery.tables.updateData",
    "pubsub.topics.publish",
    "secretmanager.versions.access",
    "workflows.executions.create",
    "workflows.executions.get",
  ]
  members    = ["serviceAccount:${local.service_account_id}@${var.project_id}.iam.gserviceaccount.com"]
  depends_on = [google_service_account.remora_service_account]
}

module "recs_workflow_custom_org_role" {
  source       = "terraform-google-modules/iam/google//modules/custom_role_iam"
  version      = "7.4.0"
  target_level = "org"
  stage        = "ALPHA"
  target_id    = var.organization_id
  role_id      = format("recs_workflow_org_sa_permissions_%s", random_id.role_id_suffix.hex)
  title        = "recs_workflow_org_sa_permissions"
  description  = "Org level IAM permissions for the Service Account to run the Recommendations workflow."
  permissions = [
    "cloudasset.assets.searchAllIamPolicies",
    "recommender.resourcemanagerProjectUtilizationInsights.get",
    "recommender.resourcemanagerProjectUtilizationInsights.list",
    "recommender.resourcemanagerProjectUtilizationRecommendations.get",
    "recommender.resourcemanagerProjectUtilizationRecommendations.list",
    "recommender.resourcemanagerProjectUtilizationRecommendations.update",
    "resourcemanager.projects.delete",
    "resourcemanager.projects.get",
    "resourcemanager.projects.getIamPolicy",
    "resourcemanager.projects.list",
    "resourcemanager.projects.update",
  ]
  members    = ["serviceAccount:${local.service_account_id}@${var.project_id}.iam.gserviceaccount.com"]
  depends_on = [google_service_account.remora_service_account]
}

# Workflows Common
resource "google_workflows_workflow" "recommendations_workflow_main" {
  name            = "recommendations_workflow_main"
  region          = var.region
  service_account = google_service_account.remora_service_account.email
  source_contents = file("${path.module}/workflows/recommendations_workflow_main.yaml")
}

resource "google_workflows_workflow" "recommendations_workflow_initial_setup" {
  name            = "recommendations_workflow_initial_setup"
  region          = var.region
  service_account = google_service_account.remora_service_account.email
  source_contents = file("${path.module}/workflows/recommendations_workflow_initial_setup.yaml")
}

resource "google_workflows_workflow" "recommendations_workflow_process_recommendations" {
  name            = "recommendations_workflow_process_recommendations"
  region          = var.region
  service_account = google_service_account.remora_service_account.email
  source_contents = file("${path.module}/workflows/recommendations_workflow_process_recommendations.yaml")
}

resource "google_workflows_workflow" "recommendations_workflow_summarize_and_notify" {
  name            = "recommendations_workflow_summarize_and_notify"
  region          = var.region
  service_account = google_service_account.remora_service_account.email
  source_contents = file("${path.module}/workflows/recommendations_workflow_summarize_and_notify.yaml")
}

# Pub Sub topic and subscription

resource "google_pubsub_topic" "recommendations_workflow" {
  name = "recommendations-workflow-topic"
}

resource "google_pubsub_subscription" "recommendations_workflow" {
  name  = "recommendations-workflow-subscription"
  topic = google_pubsub_topic.recommendations_workflow.name
}

# Cloud Functions

resource "google_storage_bucket" "recommendation_workflow_functions" {
  name                        = "${var.project_id}-recommendations-workflow-functions"
  location                    = "US"
  uniform_bucket_level_access = true
}

data "archive_file" "recommendation_workflow_function" {
  type        = "zip"
  output_path = "/tmp/recommendation-workflow.zip"
  source_dir  = "${path.module}/functions/${var.cloudfunction_notifier}"
}

resource "google_storage_bucket_object" "recommendation_workflow_function_zip" {
  name   = "recommendation-workflow-function.zip"
  bucket = google_storage_bucket.recommendation_workflow_functions.id
  source = data.archive_file.recommendation_workflow_function.output_path
}

resource "google_cloudfunctions_function" "send_email" {
  count                 = var.cloudfunction_notifier == "sendgrid" ? 1 : 0
  name                  = "send-sendgrid-email"
  description           = "Sends email notification via Sendgrid for Recommendations Workflow"
  runtime               = "python38"
  region                = var.region
  service_account_email = google_service_account.remora_service_account.email
  entry_point           = "send_email"
  event_trigger {
    event_type = "google.pubsub.topic.publish"
    resource   = google_pubsub_topic.recommendations_workflow.id
  }
  source_archive_bucket = google_storage_bucket.recommendation_workflow_functions.id
  source_archive_object = google_storage_bucket_object.recommendation_workflow_function_zip.name
  secret_environment_variables {
    key     = "sendgrid-key"
    secret  = var.sendgrid_config.sendgrid_secret_name
    version = "latest"
  }
  environment_variables = local.sendgrid_cloudfunctions_environment_variables
}

resource "google_cloudfunctions_function" "create_jira_issue" {
  count                 = var.cloudfunction_notifier == "jira" ? 1 : 0
  name                  = "create-jira-issue"
  description           = "Creates Jira tickets for Recommendations Workflow"
  runtime               = "python38"
  region                = var.region
  service_account_email = google_service_account.remora_service_account.email
  entry_point           = "create_jira_issue"
  event_trigger {
    event_type = "google.pubsub.topic.publish"
    resource   = google_pubsub_topic.recommendations_workflow.id
  }
  source_archive_bucket = google_storage_bucket.recommendation_workflow_functions.id
  source_archive_object = google_storage_bucket_object.recommendation_workflow_function_zip.name

  secret_environment_variables {
    key     = "jira-key"
    secret  = var.jira_config.jira_secret_name
    version = "latest"
  }
  environment_variables = local.jira_cloudfunctions_environment_variables
}

# Cloud Scheduler

resource "google_cloud_scheduler_job" "recommendations_workflow_run" {
  name        = "scheduled_recommendations_workflow_run"
  description = ""
  schedule    = var.schedule
  region      = var.region
  time_zone   = var.time_zone
  http_target {
    http_method = "POST"
    uri         = "https://workflowexecutions.googleapis.com/v1/projects/${var.project_id}/locations/${var.region}/workflows/${google_workflows_workflow.recommendations_workflow_main.name}/executions"
    body = base64encode(<<-EOF
      {"argument": "{\"numDaysTTL\": ${var.ttl_days}, 
      \"region\": \"${var.region}\", 
      \"timeZoneName\": \"${var.time_zone}\", 
      \"organizationId\": ${var.organization_id}, 
      \"optOutProjectNumbers\": [${join(",", var.opt_out_project_numbers)}], 
      \"essentialContactCategories\": [${join(",", var.essential_contact_categories)}], 
      \"isDryRun\": ${var.is_dry_run},
      \"allowMetrics\": ${var.allow_metrics}}"}
    EOF
    )

    oauth_token {
      service_account_email = google_service_account.remora_service_account.email
    }
  }

}
