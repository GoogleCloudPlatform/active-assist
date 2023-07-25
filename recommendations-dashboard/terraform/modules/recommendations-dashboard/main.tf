# Copyright 2023 Google LLC
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

#
# Provider
terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 4.0.0"
    }
  }

  provider_meta "google" {
    module_name = "cloud-solutions/recommendations-dashboard-v1.0"
  }
}


#
# Service account
resource "google_service_account" "rec_dashboard_sa" {
  account_id   = "rec-dashboard-sa"
  display_name = "rec-dashboard-service-account"
  project      = var.project_id
}

resource "google_organization_iam_member" "cloudasset_viewer" {
  org_id  = var.organization_id
  role    = "roles/cloudasset.viewer"
  member  = "serviceAccount:${google_service_account.rec_dashboard_sa.email}"
}

resource "google_organization_iam_member" "recommendation_exporter" {
  org_id  = var.organization_id
  role    = "roles/recommender.exporter"
  member  = "serviceAccount:${google_service_account.rec_dashboard_sa.email}"
}

resource "google_project_iam_member" "project_editor" {
  project = var.project_id
  role    = "roles/editor"
  member  = "serviceAccount:${google_service_account.rec_dashboard_sa.email}"
}

resource "google_project_iam_member" "service_usage_admin" {
  project = var.project_id
  role    = "roles/serviceusage.serviceUsageAdmin"
  member  = "serviceAccount:${google_service_account.rec_dashboard_sa.email}"
}

resource "google_bigquery_dataset_iam_member" "bq_dataset_owner" {
  dataset_id = google_bigquery_dataset.rec_dashboard_dataset.dataset_id
  role       = "roles/bigquery.dataOwner"
  member     = "serviceAccount:${google_service_account.rec_dashboard_sa.email}"
}


#
# BigQuery resources
resource "google_bigquery_dataset" "rec_dashboard_dataset" {
  dataset_id                 = "rec_dashboard_dataset"
  delete_contents_on_destroy = false
  location                   = var.bq_dataset_location
  project                    = var.project_id
}

resource "google_bigquery_table" "recommendations_export" {
  dataset_id  = google_bigquery_dataset.rec_dashboard_dataset.dataset_id
  description = "Recommendations and Insights exported by organization"
  project     = var.project_id
  schema     = file("${path.module}/recommendations_export_schema.sql")
  table_id    = "recommendations_export"

  time_partitioning {
    type = "DAY"
  }
}

resource "google_bigquery_table" "insights_export" {
  dataset_id  = google_bigquery_dataset.rec_dashboard_dataset.dataset_id
  description = "Recommendations and Insights exported by organization"
  project     = var.project_id
  schema     = file("${path.module}/insights_export_schema.sql")
  table_id    = "insights_export"

  time_partitioning {
    type = "DAY"
  }
}

resource "google_bigquery_table" "asset_export_table" {
  dataset_id = google_bigquery_dataset.rec_dashboard_dataset.dataset_id
  project    = var.project_id
  table_id   = "asset_export_table"
  schema     = file("${path.module}/asset_export_table_schema.sql")
}

resource "google_bigquery_table" "flattened_recommendations" {
  dataset_id = google_bigquery_dataset.rec_dashboard_dataset.dataset_id
  project    = var.project_id
  table_id   = "flattened_recommendations"

  view {
    query = templatefile("${path.module}/flattened_recommendations_view.sql.tftpl", {
      recommendations_export_table = "${var.project_id}.${google_bigquery_dataset.rec_dashboard_dataset.dataset_id}.${google_bigquery_table.recommendations_export.table_id}",
      asset_export_table = "${var.project_id}.${google_bigquery_dataset.rec_dashboard_dataset.dataset_id}.${google_bigquery_table.asset_export_table.table_id}",
      insights_export_table = "${var.project_id}.${google_bigquery_dataset.rec_dashboard_dataset.dataset_id}.${google_bigquery_table.insights_export.table_id}",
    })
    use_legacy_sql = false
  }
}

resource "google_bigquery_table" "flattened_cost_only_no_resource_duplicates" {
  dataset_id = google_bigquery_dataset.rec_dashboard_dataset.dataset_id
  project    = var.project_id
  table_id   = "flattened_cost_only_no_resource_duplicates"

  view {
    query = templatefile("${path.module}/flattened_cost_only_no_resource_duplicates_view.sql.tftpl", {
      flattened_recommendations_table = "${var.project_id}.${google_bigquery_dataset.rec_dashboard_dataset.dataset_id}.${google_bigquery_table.flattened_recommendations.table_id}",
    })
    use_legacy_sql = false
  }
}

resource "google_bigquery_table" "exports_data_by_week" {
  dataset_id = google_bigquery_dataset.rec_dashboard_dataset.dataset_id
  project    = var.project_id
  table_id   = "exports_data_by_week"

  view {
    query = templatefile("${path.module}/exports_data_by_week_view.sql.tftpl", {
      recommendations_export_table = "${var.project_id}.${google_bigquery_dataset.rec_dashboard_dataset.dataset_id}.${google_bigquery_table.recommendations_export.table_id}",
      asset_export_table = "${var.project_id}.${google_bigquery_dataset.rec_dashboard_dataset.dataset_id}.${google_bigquery_table.asset_export_table.table_id}",
    })
    use_legacy_sql = false
  }
}

#
# Workflows
resource "google_workflows_workflow" "rec_dashboard_workflow_main" {
  name            = "rec_dashboard_workflow_main"
  region          = var.region
  project         = var.project_id
  service_account = google_service_account.rec_dashboard_sa.email
  source_contents = file("${path.module}/workflows/recommender-api-export-workflow.yaml")
}

#
# Cloud Scheduler
resource "google_cloud_scheduler_job" "rec_dashboard_workflow_run" {
  name        = "scheduled_rec_dashboard_workflow_run"
  description = ""
  schedule    = var.schedule
  region      = var.region
  time_zone   = var.time_zone
  http_target {
    http_method = "POST"
    uri         = "https://workflowexecutions.googleapis.com/v1/${google_workflows_workflow.rec_dashboard_workflow_main.id}/executions"
    body = base64encode(jsonencode({
      "argument": jsonencode(merge ({
                "assetTable" : "${google_bigquery_table.asset_export_table.table_id}",
                "bqLocation" : "${var.bq_dataset_location}",
                "datasetId" : "${google_bigquery_dataset.rec_dashboard_dataset.dataset_id}",
                "orgId" : "${var.organization_id}",
                "projectId" : "${var.project_id}",
                "recommendationTable" : "${google_bigquery_table.recommendations_export.table_id}"
        }))
    }))

    oauth_token {
      service_account_email = google_service_account.rec_dashboard_sa.email
    }
  }
}
