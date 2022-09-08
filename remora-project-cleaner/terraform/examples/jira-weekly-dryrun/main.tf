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

module "recommendations_workflow_jira" {
  source                 = "../../modules/project-cleanup"
  project_id             = var.project_id
  organization_id        = var.organization_id
  cloudfunction_notifier = "jira"

  jira_config = {
    jira_secret_name         = var.jira_secret_name
    jira_admin_email         = var.jira_admin_email
    jira_server_address      = var.jira_server_address
    jira_project_id          = var.jira_project_id
    jira_assignee_account_id = var.jira_assignee_account_id
  }
  time_zone     = "America/Los_Angeles"
  is_dry_run    = true
  schedule      = "0 0 * * 0"
  enable_apis   = var.enable_apis
  allow_metrics = var.allow_metrics
}
