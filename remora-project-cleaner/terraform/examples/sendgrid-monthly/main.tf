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

module "recommendations_workflow_sendgrid" {
  source                 = "../../modules/project-cleanup"
  project_id             = var.project_id
  organization_id        = var.organization_id
  cloudfunction_notifier = "sendgrid"
  sendgrid_config = {
    sendgrid_secret_name    = var.sendgrid_secret_name
    sendgrid_reply_to_email = var.sendgrid_reply_to_email
    sendgrid_sender_email   = var.sendgrid_sender_email
  }
  time_zone   = "America/Los_Angeles"
  is_dry_run  = false
  schedule    = "0 0 1 * *"
  enable_apis = var.enable_apis
  ttl_days    = "60"
  allow_metrics = var.allow_metrics
}
