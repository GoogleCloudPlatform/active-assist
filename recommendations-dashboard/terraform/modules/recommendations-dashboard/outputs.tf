/**
 * Copyright 2020 Google LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

output "dashboard_template_url" {
    value = "https://lookerstudio.google.com/reporting/create?c.reportId=ae7a511f-fc98-4364-bad9-adf44a4574e7&ds.ds13.connector=bigQuery&ds.ds13.type=TABLE&ds.ds13.keepDatasourceName=true&ds.ds13.projectId=${var.project_id}&ds.ds13.datasetId=${google_bigquery_dataset.rec_dashboard_dataset.dataset_id}&ds.ds13.tableId=${google_bigquery_table.flattened_recommendations.table_id}&ds.ds24.connector=bigQuery&ds.ds24.type=TABLE&ds.ds24.keepDatasourceName=true&ds.ds24.projectId=${var.project_id}&ds.ds24.datasetId=${google_bigquery_dataset.rec_dashboard_dataset.dataset_id}&ds.ds24.tableId=${google_bigquery_table.exports_data_by_week.table_id}&ds.ds32.connector=bigQuery&ds.ds32.type=TABLE&ds.ds32.keepDatasourceName=true&ds.ds32.projectId=${var.project_id}&ds.ds32.datasetId=${google_bigquery_dataset.rec_dashboard_dataset.dataset_id}&ds.ds32.tableId=${google_bigquery_table.flattened_cost_only_no_resource_duplicates.table_id}"
}