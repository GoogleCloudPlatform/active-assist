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
# Service account
resource "google_service_account" "org_level_rec_hub_sa" {
  account_id   = "org-level-rec-hub-sa"
  display_name = "org-level-rec-hub-service-account"
  project      = var.project_id
}

resource "google_organization_iam_member" "cloudasset_viewer" {
  org_id  = var.organization_id
  role    = "roles/cloudasset.viewer"
  member  = "serviceAccount:${google_service_account.org_level_rec_hub_sa.email}"
}

resource "google_organization_iam_member" "recommendation_exporter" {
  org_id  = var.organization_id
  role    = "roles/recommender.exporter"
  member  = "serviceAccount:${google_service_account.org_level_rec_hub_sa.email}"
}

resource "google_project_iam_member" "project_editor" {
  project = var.project_id
  role    = "roles/editor"
  member  = "serviceAccount:${google_service_account.org_level_rec_hub_sa.email}"
}

resource "google_project_iam_member" "service_usage_admin" {
  project = var.project_id
  role    = "roles/serviceusage.serviceUsageAdmin"
  member  = "serviceAccount:${google_service_account.org_level_rec_hub_sa.email}"
}

resource "google_bigquery_dataset_iam_member" "bq_dataset_owner" {
  dataset_id = google_bigquery_dataset.org_level_rec_hub_dataset.dataset_id
  role       = "roles/bigquery.dataOwner"
  member     = "serviceAccount:${google_service_account.org_level_rec_hub_sa.email}"
}


#
# BigQuery resources
resource "google_bigquery_dataset" "org_level_rec_hub_dataset" {
  dataset_id                 = "org_level_rec_hub_dataset"
  delete_contents_on_destroy = false
  location                   = var.bq_dataset_location
  project                    = var.project_id
}

resource "google_bigquery_table" "recommendations_export" {
  dataset_id  = google_bigquery_dataset.org_level_rec_hub_dataset.dataset_id
  description = "Recommendations and Insights exported by organization"
  project     = var.project_id
  schema      = "[{\"description\":\"Represents what cloud entity type the recommendation was generated for - eg: project number, billing account\\n\",\"mode\":\"NULLABLE\",\"name\":\"cloud_entity_type\",\"type\":\"STRING\"},{\"description\":\"Value of the project number or billing account id\\n\",\"mode\":\"NULLABLE\",\"name\":\"cloud_entity_id\",\"type\":\"STRING\"},{\"description\":\"Name of recommendation. A project recommendation is represented as\\nprojects/[PROJECT_NUMBER]/locations/[LOCATION]/recommenders/[RECOMMENDER_ID]/recommendations/[RECOMMENDATION_ID]\\n\",\"mode\":\"NULLABLE\",\"name\":\"name\",\"type\":\"STRING\"},{\"description\":\"Location for which this recommendation is generated\\n\",\"mode\":\"NULLABLE\",\"name\":\"location\",\"type\":\"STRING\"},{\"description\":\"Recommender ID of the recommender that has produced this recommendation\\n\",\"mode\":\"NULLABLE\",\"name\":\"recommender\",\"type\":\"STRING\"},{\"description\":\"Contains an identifier for a subtype of recommendations produced for the\\nsame recommender. Subtype is a function of content and impact, meaning a\\nnew subtype will be added when either content or primary impact category\\nchanges.\\nExamples:\\nFor recommender = \\\"google.iam.policy.Recommender\\\",\\nrecommender_subtype can be one of \\\"REMOVE_ROLE\\\"/\\\"REPLACE_ROLE\\\"\\n\",\"mode\":\"NULLABLE\",\"name\":\"recommender_subtype\",\"type\":\"STRING\"},{\"description\":\"Contains the fully qualified resource names for resources changed by the\\noperations in this recommendation. This field is always populated. ex:\\n[//cloudresourcemanager.googleapis.com/projects/foo].\\n\",\"mode\":\"REPEATED\",\"name\":\"target_resources\",\"type\":\"STRING\"},{\"description\":\"Required. Free-form human readable summary in English.\\nThe maximum length is 500 characters.\\n\",\"mode\":\"NULLABLE\",\"name\":\"description\",\"type\":\"STRING\"},{\"description\":\"Output only. Last time this recommendation was refreshed by the system that created it in the first place.\\n\",\"mode\":\"NULLABLE\",\"name\":\"last_refresh_time\",\"type\":\"TIMESTAMP\"},{\"description\":\"Required. The primary impact that this recommendation can have while trying to optimize\\nfor one category.\\n\",\"fields\":[{\"description\":\"Category that is being targeted.\\nValues can be the following:\\n  CATEGORY_UNSPECIFIED:\\n    Default unspecified category. Don't use directly.\\n  COST:\\n    Indicates a potential increase or decrease in cost.\\n  SECURITY:\\n    Indicates a potential increase or decrease in security.\\n  PERFORMANCE:\\n    Indicates a potential increase or decrease in performance.\\n\",\"mode\":\"NULLABLE\",\"name\":\"category\",\"type\":\"STRING\"},{\"description\":\"Optional. Use with CategoryType.COST\",\"fields\":[{\"description\":\"An approximate projection on amount saved or amount incurred.\\nNegative cost units indicate cost savings and positive cost units indicate\\nincrease. See google.type.Money documentation for positive/negative units.\\n\",\"fields\":[{\"description\":\"The 3-letter currency code defined in ISO 4217.\",\"mode\":\"NULLABLE\",\"name\":\"currency_code\",\"type\":\"STRING\"},{\"description\":\"The whole units of the amount. For example if `currencyCode` is `\\\"USD\\\"`,\\nthen 1 unit is one US dollar.\\n\",\"mode\":\"NULLABLE\",\"name\":\"units\",\"type\":\"INTEGER\"},{\"description\":\"Number of nano (10^-9) units of the amount.\\nThe value must be between -999,999,999 and +999,999,999 inclusive.\\nIf `units` is positive, `nanos` must be positive or zero.\\nIf `units` is zero, `nanos` can be positive, zero, or negative.\\nIf `units` is negative, `nanos` must be negative or zero.\\nFor example $-1.75 is represented as `units`=-1 and `nanos`=-750,000,000.\\n\",\"mode\":\"NULLABLE\",\"name\":\"nanos\",\"type\":\"INTEGER\"}],\"mode\":\"NULLABLE\",\"name\":\"cost\",\"type\":\"RECORD\"},{\"description\":\"Duration for which this cost applies.\",\"fields\":[{\"description\":\"Signed seconds of the span of time. Must be from -315,576,000,000\\nto +315,576,000,000 inclusive. Note: these bounds are computed from:\\n60 sec/min * 60 min/hr * 24 hr/day * 365.25 days/year * 10000 years\\n\",\"mode\":\"NULLABLE\",\"name\":\"seconds\",\"type\":\"INTEGER\"},{\"description\":\"Signed fractions of a second at nanosecond resolution of the span\\nof time. Durations less than one second are represented with a 0\\n`seconds` field and a positive or negative `nanos` field. For durations\\nof one second or more, a non-zero value for the `nanos` field must be\\nof the same sign as the `seconds` field. Must be from -999,999,999\\nto +999,999,999 inclusive.\\n\",\"mode\":\"NULLABLE\",\"name\":\"nanos\",\"type\":\"INTEGER\"}],\"mode\":\"NULLABLE\",\"name\":\"duration\",\"type\":\"RECORD\"}],\"mode\":\"NULLABLE\",\"name\":\"cost_projection\",\"type\":\"RECORD\"}],\"mode\":\"NULLABLE\",\"name\":\"primary_impact\",\"type\":\"RECORD\"},{\"description\":\"Output only. The state of the recommendation:\\n  STATE_UNSPECIFIED:\\n    Default state. Don't use directly.\\n  ACTIVE:\\n    Recommendation is active and can be applied. Recommendations content can\\n    be updated by Google.\\n    ACTIVE recommendations can be marked as CLAIMED, SUCCEEDED, or FAILED.\\n  CLAIMED:\\n    Recommendation is in claimed state. Recommendations content is\\n    immutable and cannot be updated by Google.\\n    CLAIMED recommendations can be marked as CLAIMED, SUCCEEDED, or FAILED.\\n  SUCCEEDED:\\n    Recommendation is in succeeded state. Recommendations content is\\n    immutable and cannot be updated by Google.\\n    SUCCEEDED recommendations can be marked as SUCCEEDED, or FAILED.\\n  FAILED:\\n    Recommendation is in failed state. Recommendations content is immutable\\n    and cannot be updated by Google.\\n    FAILED recommendations can be marked as SUCCEEDED, or FAILED.\\n  DISMISSED:\\n    Recommendation is in dismissed state.\\n    DISMISSED recommendations can be marked as ACTIVE.\\n\",\"mode\":\"NULLABLE\",\"name\":\"state\",\"type\":\"STRING\"},{\"description\":\"Ancestry for the recommendation entity\\n\",\"fields\":[{\"description\":\"Organization to which the recommendation project\\n\",\"mode\":\"NULLABLE\",\"name\":\"organization_id\",\"type\":\"STRING\"},{\"description\":\"Up to 5 levels of parent folders for the recommendation project\\n\",\"mode\":\"REPEATED\",\"name\":\"folder_ids\",\"type\":\"STRING\"}],\"mode\":\"NULLABLE\",\"name\":\"ancestors\",\"type\":\"RECORD\"},{\"description\":\"Insights associated with this recommendation. A project insight is represented as\\nprojects/[PROJECT_NUMBER]/locations/[LOCATION]/insightTypes/[INSIGHT_TYPE_ID]/insights/[insight_id]\\n\",\"mode\":\"REPEATED\",\"name\":\"associated_insights\",\"type\":\"STRING\"},{\"description\":\"Additional details about the recommendation in JSON format\\n\",\"mode\":\"NULLABLE\",\"name\":\"recommendation_details\",\"type\":\"STRING\"},{\"description\":\"Priority of the recommendation:\\n  PRIORITY_UNSPECIFIED:\\n    Default unspecified priority. Don't use directly.\\n  P4:\\n    Lowest priority.\\n  P3:\\n    Second lowest priority.\\n  P2:\\n    Second highest priority.\\n  P1:\\n    Highest priority.\\n\",\"mode\":\"NULLABLE\",\"name\":\"priority\",\"type\":\"STRING\"}]"
  table_id    = "recommendations_export"

  time_partitioning {
    type = "DAY"
  }
}

resource "google_bigquery_table" "insights_export" {
  dataset_id  = google_bigquery_dataset.org_level_rec_hub_dataset.dataset_id
  description = "Recommendations and Insights exported by organization"
  project     = var.project_id
  schema      = "[{\"description\":\"Represents what cloud entity type the recommendation was generated for - eg: project number, billing account\\n\",\"mode\":\"NULLABLE\",\"name\":\"cloud_entity_type\",\"type\":\"STRING\"},{\"description\":\"Value of the project number or billing account id\\n\",\"mode\":\"NULLABLE\",\"name\":\"cloud_entity_id\",\"type\":\"STRING\"},{\"description\":\"Name of recommendation. A project recommendation is represented as\\nprojects/[PROJECT_NUMBER]/locations/[LOCATION]/recommenders/[RECOMMENDER_ID]/recommendations/[RECOMMENDATION_ID]\\n\",\"mode\":\"NULLABLE\",\"name\":\"name\",\"type\":\"STRING\"},{\"description\":\"Location for which this recommendation is generated\\n\",\"mode\":\"NULLABLE\",\"name\":\"location\",\"type\":\"STRING\"},{\"description\":\"Recommender ID of the recommender that has produced this recommendation\\n\",\"mode\":\"NULLABLE\",\"name\":\"insight_type\",\"type\":\"STRING\"},{\"description\":\"Contains an identifier for a subtype of recommendations produced for the\\nsame recommender. Subtype is a function of content and impact, meaning a\\nnew subtype will be added when either content or primary impact category\\nchanges.\\nExamples:\\nFor recommender = \\\"google.iam.policy.Recommender\\\",\\nrecommender_subtype can be one of \\\"REMOVE_ROLE\\\"/\\\"REPLACE_ROLE\\\"\\n\",\"mode\":\"NULLABLE\",\"name\":\"insight_subtype\",\"type\":\"STRING\"},{\"description\":\"Contains the fully qualified resource names for resources changed by the\\noperations in this recommendation. This field is always populated. ex:\\n[//cloudresourcemanager.googleapis.com/projects/foo].\\n\",\"mode\":\"REPEATED\",\"name\":\"target_resources\",\"type\":\"STRING\"},{\"description\":\"Required. Free-form human readable summary in English.\\nThe maximum length is 500 characters.\\n\",\"mode\":\"NULLABLE\",\"name\":\"description\",\"type\":\"STRING\"},{\"description\":\"Output only. Last time this recommendation was refreshed by the system that created it in the first place.\\n\",\"mode\":\"NULLABLE\",\"name\":\"last_refresh_time\",\"type\":\"TIMESTAMP\"},{\"description\":\"Category being targeted by the insight. Can be one of:\\nUnspecified category.\\nCATEGORY_UNSPECIFIED = Unspecified category.\\nCOST = The insight is related to cost.\\nSECURITY = The insight is related to security.\\nPERFORMANCE = The insight is related to performance.\\nMANAGEABILITY = The insight is related to manageability.;\\n\",\"mode\":\"NULLABLE\",\"name\":\"category\",\"type\":\"STRING\"},{\"description\":\"Output only. The state of the recommendation:\\n  STATE_UNSPECIFIED:\\n    Default state. Don't use directly.\\n  ACTIVE:\\n    Recommendation is active and can be applied. Recommendations content can\\n    be updated by Google.\\n    ACTIVE recommendations can be marked as CLAIMED, SUCCEEDED, or FAILED.\\n  CLAIMED:\\n    Recommendation is in claimed state. Recommendations content is\\n    immutable and cannot be updated by Google.\\n    CLAIMED recommendations can be marked as CLAIMED, SUCCEEDED, or FAILED.\\n  SUCCEEDED:\\n    Recommendation is in succeeded state. Recommendations content is\\n    immutable and cannot be updated by Google.\\n    SUCCEEDED recommendations can be marked as SUCCEEDED, or FAILED.\\n  FAILED:\\n    Recommendation is in failed state. Recommendations content is immutable\\n    and cannot be updated by Google.\\n    FAILED recommendations can be marked as SUCCEEDED, or FAILED.\\n  DISMISSED:\\n    Recommendation is in dismissed state.\\n    DISMISSED recommendations can be marked as ACTIVE.\\n\",\"mode\":\"NULLABLE\",\"name\":\"state\",\"type\":\"STRING\"},{\"description\":\"Ancestry for the recommendation entity\\n\",\"fields\":[{\"description\":\"Organization to which the recommendation project\\n\",\"mode\":\"NULLABLE\",\"name\":\"organization_id\",\"type\":\"STRING\"},{\"description\":\"Up to 5 levels of parent folders for the recommendation project\\n\",\"mode\":\"REPEATED\",\"name\":\"folder_ids\",\"type\":\"STRING\"}],\"mode\":\"NULLABLE\",\"name\":\"ancestors\",\"type\":\"RECORD\"},{\"description\":\"Insights associated with this recommendation. A project insight is represented as\\nprojects/[PROJECT_NUMBER]/locations/[LOCATION]/insightTypes/[INSIGHT_TYPE_ID]/insights/[insight_id]\\n\",\"mode\":\"REPEATED\",\"name\":\"associated_recommendations\",\"type\":\"STRING\"},{\"description\":\"Additional details about the insight in JSON format\\nschema:\\n  fields:\\n  - name: content\\n    type: STRING\\n    description: |\\n      A struct of custom fields to explain the insight.\\n      Example: \\\"grantedPermissionsCount\\\": \\\"1000\\\"\\n  - name: observation_period\\n    type: TIMESTAMP\\n    description: |\\n      Observation period that led to the insight. The source data used to\\n      generate the insight ends at last_refresh_time and begins at\\n      (last_refresh_time - observation_period).\\n- name: state_metadata\\n  type: STRING\\n  description: |\\n    A map of metadata for the state, provided by user or automations systems.\\n\",\"mode\":\"NULLABLE\",\"name\":\"insight_details\",\"type\":\"STRING\"},{\"description\":\"Severity of the insight:\\n  SEVERITY_UNSPECIFIED:\\n    Default unspecified severity. Don't use directly.\\n  LOW:\\n    Lowest severity.\\n  MEDIUM:\\n    Second lowest severity.\\n  HIGH:\\n    Second highest severity.\\n  CRITICAL:\\n    Highest severity.\\n\",\"mode\":\"NULLABLE\",\"name\":\"severity\",\"type\":\"STRING\"}]"
  table_id    = "insights_export"

  time_partitioning {
    type = "DAY"
  }
}

resource "google_bigquery_table" "asset_export_table" {
  dataset_id = google_bigquery_dataset.org_level_rec_hub_dataset.dataset_id
  project    = var.project_id
  table_id   = "asset_export_table"
  schema     = <<EOF
    [
      {"mode":"NULLABLE","name":"name","type":"STRING"},
      {"mode":"NULLABLE","name":"asset_type","type":"STRING"},
      {"mode":"NULLABLE","name":"resource","type":"RECORD", "fields":[
        {"mode":"NULLABLE","name":"version","type":"STRING"},
        {"mode":"NULLABLE","name":"discovery_document_uri","type":"STRING"},
        {"mode":"NULLABLE","name":"discovery_name","type":"STRING"},
        {"mode":"NULLABLE","name":"resource_url","type":"STRING"},
        {"mode":"NULLABLE","name":"parent","type":"STRING"},
        {"mode":"NULLABLE","name":"data","type":"STRING"},
        {"mode":"NULLABLE","name":"location","type":"STRING"}
      ]},
      {"mode":"REPEATED","name":"ancestors","type":"STRING"},
      {"mode":"NULLABLE","name":"update_time","type":"TIMESTAMP"}
    ]
  EOF
}

resource "google_bigquery_table" "flattened_recommendations" {
  dataset_id = google_bigquery_dataset.org_level_rec_hub_dataset.dataset_id
  project    = var.project_id
  table_id   = "flattened_recommendations"
  schema     = <<EOF
    [
      {"mode":"NULLABLE","name":"project_name","type":"STRING"},
      {"mode":"NULLABLE","name":"project_id","type":"STRING"},
      {"mode":"NULLABLE","name":"recommender_name","type":"STRING"},
      {"mode":"NULLABLE","name":"location","type":"STRING"},
      {"mode":"NULLABLE","name":"recommender_subtype","type":"STRING"},
      {"mode":"REPEATED","name":"target_resources","type":"STRING"},
      {"mode":"NULLABLE","name":"recommender_last_refresh_time","type":"TIMESTAMP"},
      {"mode":"NULLABLE","name":"impact_category","type":"STRING"},
      {"mode":"NULLABLE","name":"has_impact_cost","type":"BOOLEAN"},
      {"mode":"NULLABLE","name":"impact_cost_unit","type":"INTEGER"},
      {"mode":"NULLABLE","name":"impact_currency_code","type":"STRING"},
      {"mode":"NULLABLE","name":"recommender_state","type":"STRING"},
      {"mode":"NULLABLE","name":"description","type":"STRING"},
      {"mode":"REPEATED","name":"folder_ids","type":"STRING"},
      {"mode":"REPEATED","name":"insight_ids","type":"STRING"},
      {"mode":"REPEATED","name":"insights","type":"RECORD", "fields":[
        {"mode":"NULLABLE","name":"insight_name","type":"STRING"},
        {"mode":"NULLABLE","name":"insight_type","type":"STRING"},
        {"mode":"NULLABLE","name":"insight_subtype","type":"STRING"},
        {"mode":"NULLABLE","name":"insight_category","type":"STRING"},
        {"mode":"NULLABLE","name":"insight_state","type":"STRING"}
      ]}
    ]
  EOF

  view {
    query          = <<EOF
      select 
        project_name,
        project_id,
        name as recommender_name,
        location,
        REPLACE(recommender_subtype, "_", " ") as recommender_subtype,
        ARRAY_AGG(distinct target_resource) as target_resources,
        last_refresh_time as recommender_last_refresh_time,
        primary_impact.category as impact_category,
        if(primary_impact.cost_projection.cost.units is null,false, true) as has_impact_cost,
        ABS(primary_impact.cost_projection.cost.units) as impact_cost_unit,
        primary_impact.cost_projection.cost.currency_code as impact_currency_code,
        state as recommender_state,
        description,
        ARRAY_AGG(distinct folder_id ignore nulls) as folder_ids,
        ARRAY_AGG(distinct insight_id) as insight_ids,
        ARRAY_AGG(STRUCT(insight_name, insight_type, insight_subtype, category as insight_category, insight_state)) as insights
        from (
          select * except(associated_insights, target_resources)
          # We just want to grab the latest refresh time per export
          from (SELECT
              agg.table.*
            FROM (
              select target_resource,
              recommender_subtype,
                  ARRAY_AGG(STRUCT(table)
                    ORDER BY  
                      last_refresh_time DESC)[SAFE_OFFSET(0)] agg
              FROM
                `${var.project_id}.${google_bigquery_dataset.org_level_rec_hub_dataset.dataset_id}.${google_bigquery_table.recommendations_export.table_id}` table cross join unnest(target_resources) as target_resource
                GROUP BY target_resource, recommender_subtype)
          ) as r
          cross join unnest(associated_insights) as insight_id
          cross join unnest(target_resources) as target_resource
          #Cross join will remove nulls, and in our case we still need nulls
          left join unnest(ancestors.folder_ids) as folder_id
          left join (Select project_name, project_id from
            (
              select 
              if(
                    ENDS_WITH(name, "/billingInfo"),
                    REGEXP_EXTRACT(REPLACE(name,"/billingInfo",""), r'/([^/]+)/?$'),
                    REGEXP_EXTRACT(name, r'/([^/]+)/?$')
              ) as project_name,
              REGEXP_EXTRACT(ancestor,  r'/([^/]+)/?$') as project_id,
              asset_type from 
                (select * from `${var.project_id}.${google_bigquery_dataset.org_level_rec_hub_dataset.dataset_id}.${google_bigquery_table.asset_export_table.table_id}`
                cross join unnest(ancestors) as ancestor
                where asset_type in ("compute.googleapis.com/Project", "cloudbilling.googleapis.com/ProjectBillingInfo")
                and ancestor like "projects/%")
            )
          ) as a
          on r.cloud_entity_id = a.project_id
          left join (select 
            name as insight_name,
            insight_type,
            insight_subtype,
            category,
            state as insight_state,
            a_r
            from `${var.project_id}.${google_bigquery_dataset.org_level_rec_hub_dataset.dataset_id}.${google_bigquery_table.insights_export.table_id}`
            cross join unnest(associated_recommendations) as a_r
          ) as i
          on r.name=i.a_r
        )
        group by 1,2,3,4,5,7,8,9,10,11,12,13
        order by recommender_name
    EOF
    use_legacy_sql = false
  }
}

resource "google_bigquery_table" "flattened_cost_only_no_resource_duplicates" {
  dataset_id = google_bigquery_dataset.org_level_rec_hub_dataset.dataset_id
  project    = var.project_id
  schema     = <<EOF
    [
      {"mode":"NULLABLE","name":"project_name","type":"STRING"},
      {"mode":"NULLABLE","name":"project_id","type":"STRING"},
      {"mode":"NULLABLE","name":"recommender_name","type":"STRING"},
      {"mode":"NULLABLE","name":"location","type":"STRING"},
      {"mode":"NULLABLE","name":"recommender_subtype","type":"STRING"},
      {"mode":"REPEATED","name":"target_resources","type":"STRING"},
      {"mode":"NULLABLE","name":"recommender_last_refresh_time","type":"TIMESTAMP"},
      {"mode":"NULLABLE","name":"impact_category","type":"STRING"},
      {"mode":"NULLABLE","name":"has_impact_cost","type":"BOOLEAN"},
      {"mode":"NULLABLE","name":"impact_cost_unit","type":"INTEGER"},
      {"mode":"NULLABLE","name":"impact_currency_code","type":"STRING"},
      {"mode":"NULLABLE","name":"recommender_state","type":"STRING"},
      {"mode":"NULLABLE","name":"description","type":"STRING"},
      {"mode":"REPEATED","name":"folder_ids","type":"STRING"},
      {"mode":"REPEATED","name":"insight_ids","type":"STRING"},
      {"mode":"REPEATED","name":"insights","type":"RECORD", "fields":[
        {"mode":"NULLABLE","name":"insight_name","type":"STRING"},
        {"mode":"NULLABLE","name":"insight_type","type":"STRING"},
        {"mode":"NULLABLE","name":"insight_subtype","type":"STRING"},
        {"mode":"NULLABLE","name":"insight_category","type":"STRING"},
        {"mode":"NULLABLE","name":"insight_state","type":"STRING"}
      ]}
    ]
  EOF
  table_id   = "flattened_cost_only_no_resource_duplicates"

  view {
    query          = <<EOF
      SELECT
        agg.table.*
          FROM (
            select target_resource,
              ARRAY_AGG(STRUCT(table)
                ORDER BY  
                  impact_cost_unit DESC)[SAFE_OFFSET(0)] agg
            FROM
              `${var.project_id}.${google_bigquery_dataset.org_level_rec_hub_dataset.dataset_id}.${google_bigquery_table.flattened_recommendations.table_id}` table cross join unnest(target_resources) as target_resource
            where has_impact_cost = true AND recommender_subtype in ('CHANGE MACHINE TYPE', 'STOP VM')
            GROUP BY target_resource)
      union all (
        SELECT * 
        from `${var.project_id}.${google_bigquery_dataset.org_level_rec_hub_dataset.dataset_id}.${google_bigquery_table.flattened_recommendations.table_id}`
        where has_impact_cost = true AND recommender_subtype NOT IN ('CHANGE MACHINE TYPE', 'STOP VM'))
    EOF
    use_legacy_sql = false
  }
}

#
# Workflows
resource "google_workflows_workflow" "org_level_rec_hub_workflow_main" {
  name            = "org_level_rec_hub_workflow_main"
  region          = var.region
  project         = var.project_id
  service_account = google_service_account.org_level_rec_hub_sa.email
  source_contents = file("${path.module}/workflows/recommender-api-export-workflow.yaml")
}

#
# Cloud Scheduler
resource "google_cloud_scheduler_job" "org_level_rec_hub_workflow_run" {
  name        = "scheduled_org_level_rec_hub_workflow_run"
  description = ""
  schedule    = var.schedule
  region      = var.region
  time_zone   = var.time_zone
  http_target {
    http_method = "POST"
    uri         = "https://workflowexecutions.googleapis.com/v1/${google_workflows_workflow.org_level_rec_hub_workflow_main.id}/executions"
    body = base64encode(jsonencode({
      "argument": jsonencode(merge ({
                "assetTable" : "${google_bigquery_table.asset_export_table.table_id}",
                "bqLocation" : "${var.bq_dataset_location}",
                "datasetId" : "${google_bigquery_dataset.org_level_rec_hub_dataset.dataset_id}",
                "orgId" : "${var.organization_id}",
                "projectId" : "${var.project_id}",
                "recommendationTable" : "${google_bigquery_table.recommendations_export.table_id}"
        }))
    }))

    oauth_token {
      service_account_email = google_service_account.org_level_rec_hub_sa.email
    }
  }
}