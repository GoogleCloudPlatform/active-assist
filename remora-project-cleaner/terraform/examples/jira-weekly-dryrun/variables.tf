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

variable "terraform_service_account" {
  type        = string
  description = "Service account that you will use to run Terraform code."
  default     = ""
}

variable "project_id" {
  type        = string
  description = "The project ID where the Recommendations Workflow will be created."
}

variable "region" {
  type        = string
  description = "Region to use for BigQuery, Cloud Workflows, functions, and scheduler."
  default     = "us-central1"
}

variable "organization_id" {
  type        = string
  description = "Organization ID where the Recommendations Workflow will run for."
}

variable "jira_secret_name" {
  type = string
}

variable "jira_admin_email" {
  type = string
}

variable "jira_server_address" {
  type = string
}

variable "jira_project_id" {
  type = string
}

variable "jira_assignee_account_id" {
  type = string
}

variable "enable_apis" {
  type        = bool
  description = "Set to true to have Terraform enable services."
  default     = true
}

variable "allow_metrics" {
  type        = bool
  description = "If true, processing information will be stored as part of the recommendation's stateMetadata to be used to improve and measure the performance of Remora. The value of this field will not impact the functionality of Remora."
}
