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

variable "project_id" {
  type        = string
  description = "The project ID where the Recommendations Workflow will be created."
}

variable "organization_id" {
  type        = string
  description = "The organization ID whose Unattended Project Recommendations will be used for the Recommendations Workflow."
}

variable "region" {
  type        = string
  description = "Region to use for BigQuery, Cloud Workflows, functions, and scheduler."
  default     = "us-central1"
}

variable "cloudfunction_notifier" {
  type        = string
  description = "The desired method of notifying project owners eminent project deletion. Supported values are 'sendgrid' and 'jira'"
  validation {
    condition     = var.cloudfunction_notifier == "sendgrid" || var.cloudfunction_notifier == "jira"
    error_message = "Supported values for 'cloudfunction_notifier' are 'sendgrid' or 'jira'."
  }
}

variable "sendgrid_config" {
  type        = map(any)
  description = "Map of values required for Sendgrid configuration: 'sendgrid_secret_name', 'sendgrid_reply_to_email','sendgrid_sender_email'"
  default = {
    sendgrid_secret_name    = ""
    sendgrid_reply_to_email = ""
    sendgrid_sender_email   = ""
  }
}

variable "jira_config" {
  type        = map(any)
  description = "Map of values required for Jira configuration: 'jira_secret_name', 'jira_admin_email', 'jira_server_address', 'jira_project_id', 'jira_assignee_account_id'"
  default = {
    jira_secret_name         = ""
    jira_admin_email         = ""
    jira_server_address      = ""
    jira_project_id          = ""
    jira_assignee_account_id = ""
  }
}

variable "schedule" {
  type        = string
  description = "Time interval defined using a unix-cron (see https://cloud.google.com/scheduler/docs/configuring/cron-job-schedules) format. This represents how frequently the recommendations workflow will run."
  default     = "0 9 * * 1"
}

variable "is_dry_run" {
  type        = bool
  description = "If true, will not delete project when it's the 3rd+ time seeing a recommendation and the TTL date has passed. If false, will delete the project when it's the 3rd+ time seeing a recommendation and the TTL date has passed."
}

variable "time_zone" {
  type        = string
  description = "The time zone database name, which will be the time zone used for formatting and for the TTL date. A list of time zone database names can be found at https://en.wikipedia.org/wiki/List_of_tz_database_time_zones. e.g. America/Los_Angeles"
}

variable "opt_out_project_numbers" {
  type        = list(number)
  description = "List of projects that won't be processed in the recommendations workflow. e.g. [111111111111, 222222222222]"
  default     = []
}

variable "ttl_days" {
  type        = string
  description = "The number of days after which an unused project can be safely removed"
  default     = "30"
}

variable "enable_apis" {
  type        = bool
  description = "Toggle to include required APIs."
  default     = false
}

variable "essential_contact_categories" {
  type        = list(string)
  description = "Categories for the Essential Contacts that should be notified during escalation. If none provided, the parent owner will used."
  default     = []
}

variable "allow_metrics" {
  type        = bool
  description = "If true, processing information will be stored as part of the recommendation's stateMetadata to be used to improve and measure the performance of Remora. The value of this field will not impact the functionality of Remora."
}
