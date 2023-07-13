# Terraform README

This Terraform configuration file deploys a set of resources in Google Cloud Platform (GCP) to set up the recommendations dashboard.

## Prerequisites
Before using this Terraform file, ensure that you have the following:

- A GCP project where you want to deploy the resources.
- Appropriate permissions and access to create resources in the GCP project.
- Terraform installed on your local machine or the environment where you plan to run Terraform commands.

### Google Cloud APIs
Ensure that the following Google Cloud APIs are enabled in your project before running the Terraform commands:

- Cloud Scheduler API
- Cloud Resource Manager API
- BigQuery API
- Workflows API
- Identity and Access Management API

To enable the APIs using the `gcloud` command-line tool, run the following commands:

```bash
gcloud services enable cloudscheduler.googleapis.com
gcloud services enable cloudresourcemanager.googleapis.com
gcloud services enable bigquery.googleapis.com
gcloud services enable workflows.googleapis.com
gcloud services enable iam.googleapis.com
```

## Configuration

### Google Service Account
A Google service account is created with the following configuration:

- Account ID: "rec-dashboard-sa"
- Display Name: "rec-dashboard-service-account"
- Project: The project ID is provided as a variable.

### Google Organization IAM Members
Two Google Organization IAM members are created:

1. Cloud Asset Viewer:
   - Org ID: The organization ID is provided as a variable.
   - Role: "roles/cloudasset.viewer"
   - Member: "serviceAccount:${google_service_account.rec_dashboard_sa.email}"

2. Recommendation Exporter:
   - Org ID: The organization ID is provided as a variable.
   - Role: "roles/recommender.exporter"
   - Member: "serviceAccount:${google_service_account.rec_dashboard_sa.email}"

### Google Project IAM Members
Two Google Project IAM members are created:

1. Project Editor:
   - Project: The project ID is provided as a variable.
   - Role: "roles/editor"
   - Member: "serviceAccount:${google_service_account.rec_dashboard_sa.email}"

2. Service Usage Admin:
   - Project: The project ID is provided as a variable.
   - Role: "roles/serviceusage.serviceUsageAdmin"
   - Member: "serviceAccount:${google_service_account.rec_dashboard_sa.email}"

### Google BigQuery Dataset and Tables

1. Google BigQuery Dataset:
   - Dataset ID: "rec_dashboard_dataset"
   - Delete Contents on Destroy: false
   - Location: The dataset location is provided as a variable.
   - Project: The project ID is provided as a variable.

2. Google BigQuery Table - Recommendations Export:
   - Dataset ID: "rec_dashboard_dataset"
   - Description: "Recommendations and Insights exported by organization"
   - Project: The project ID is provided as a variable.
   - Schema: The schema for the table is a JSON representation defining the columns and their types.

3. Google BigQuery Table - Insights Export:
   - Dataset ID: "rec_dashboard_dataset"
   - Description: "Recommendations and Insights exported by organization"
   - Project: The project ID is provided as a variable.
   - Schema: The schema for the table is a JSON representation defining the columns and their types.

4. Google BigQuery Table - Asset Export Table:
   - Dataset ID: "rec_dashboard_dataset"
   - Project: The project ID is provided as a variable.
   - Table ID: "asset_export_table"
   - Schema: The schema for the table is a JSON representation defining the columns and their types.

5. Google BigQuery Table - Flattened Recommendations:
   - Dataset ID: "rec_dashboard_dataset

"
   - Project: The project ID is provided as a variable.
   - Table ID: "flattened_recommendations"
   - Schema: The schema for the table is a JSON representation defining the columns and their types.
   - View: A view is defined with a SQL query to retrieve specific columns and filter data.

6. Google BigQuery Table - Flattened Cost Only (No Resource Duplicates):
   - Dataset ID: "rec_dashboard_dataset"
   - Project: The project ID is provided as a variable.
   - Table ID: "flattened_cost_only_no_resource_duplicates"
   - Schema: The schema for the table is a JSON representation defining the columns and their types.
   - View: A view is defined with a SQL query to retrieve specific columns and filter data.

### Workflows
A Google Cloud Workflow is created with the following configuration:

- Name: "rec_dashboard_workflow_main"
- Region: The region is provided as a variable.
- Project: The project ID is provided as a variable.
- Service Account: The email of the previously created Google service account.
- Source Contents: The path to the YAML file defining the workflow.

### Cloud Scheduler
A Google Cloud Scheduler job is created with the following configuration:

- Name: "scheduled_rec_dashboard_workflow_run"
- Description: (Optional) Description of the job.
- Schedule: The schedule is provided as a variable.
- Region: The region is provided as a variable.
- Time Zone: The time zone is provided as a variable.
- HTTP Target: The URI and body for the HTTP target are configured to trigger the execution of the Google Cloud Workflow.
- OAuth Token: The service account email of the previously created Google service account.

Certainly! Here's an updated version of the usage section that includes the variables that need to be set:

## Usage
1. Ensure that the prerequisites are met and the required Google Cloud APIs are enabled.
2. Set the necessary input variables either directly in the Terraform file or using a separate variable file. The required variables are as follows:

   - `project_id`: The ID of the GCP project where the resources will be deployed.
   - `organization_id`: The ID of the GCP organization.
   - `bq_dataset_location`: The location of the BigQuery dataset.
   - `region`: The region where the workflows and scheduler job will be created.
   - `schedule`: The schedule for the Cloud Scheduler job in [Cron format](https://en.wikipedia.org/wiki/Cron#CRON_expression).
   - `time_zone`: The time zone for the Cloud Scheduler job.

   You can set these variables in a separate `.tfvars` file and reference it during the Terraform commands using the `-var-file` flag. For example:
   ```bash
   terraform apply -var-file=variables.tfvars
   ```

3. Initialize the Terraform working directory by running the command:
   ```bash
   terraform init
   ```

4. Preview the changes that Terraform will apply by running the command:
   ```bash
   terraform plan
   ```

5. If the preview looks satisfactory, apply the changes by running the command:
   ```bash
   terraform apply
   ```

6. Review the changes Terraform will make and confirm by entering "yes" when prompted.

7. Terraform will create the specified resources in your GCP project. Monitor the command output for any errors or additional instructions.

## Clean Up
To clean up and delete the created resources, run the following command:
```bash
terraform destroy
```
Review the planned actions and enter "yes" when prompted to proceed with the deletion.

Note: The `terraform destroy` command will remove all the resources created by this Terraform configuration.
