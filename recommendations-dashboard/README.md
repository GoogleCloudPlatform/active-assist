# Google Cloud Recommendations Dashboard

Google Cloud's [Active Assist][activeassist] tooling creates recommendations to improve the operation of your Google Cloud resources, including cost efficiency, security, and sustainability. Most recommendations are scoped and only visible at the project level. This tooling sets up exports for Recommendations and [Cloud Asset Inventory][assetinventory] to [BigQuery][bigquery] and provides [Looker Studio][lookerstudio] dashboards. The dashboards create a view into the recommendations across all the projects in your organization, with an emphasis on highlighting opportunities for cost savings.

## Architecture

![architecture](resources/architecture.png)

The diagram above show the high level components of the recommndations dashboard and how they interact.

1. Cloud Scheduler triggers a Workflow to export Cloud Asset Inventory data to BigQuery. On the initial
   execution it also setups the automatic export of Active Assist recommendations to BigQuery.

2. Once the recommendation and inventory data is in BigQuery it is deduplicated and combined using the 
   views in `sql` folder.

3. Looker Studio dashboards are then used to make the data accessible to FinOps and individual teams that
   managed Google Cloud resources.

## Deployment Guide

This section will guide you through the deployment of the Recommendations Dashboard. For more details on 
what and how things are deployed refer to documentation in  the respective folders (`sql` and `terraform`)
for detailed instructions on setting up and configuring the dashboards and infrastructure components.

### Prerequisites

1. Google Cloud Organization
2. Google Cloud Support purchased for your organization
3. A Google Cloud project that the dashboard will be deployed to

### Before you begin

In this section you prepare your project for deployment.

1.  Open the [Cloud Console][cloud-console]
2.  Activate [Cloud Shell][cloud-shell] \
    At the bottom of the Cloud Console, a
    <a href='https://cloud.google.com/shell/docs/features'>Cloud Shell</a>
    session starts and displays a command-line prompt. Cloud Shell is a shell
    environment with the Cloud SDK already installed, including the
    <code>gcloud</code> command-line tool, and with values already set for your
    current project. It can take a few seconds for the session to initialize.

3.  In Cloud Shell, clone this repository

    ```sh
    git clone https://github.com/GoogleCloudPlatform/active-assist.git
    ```

4.  Export variables for the working directories

    ```sh
    export DASHBOARD_DIR="$(pwd)/active-assist/recommendations-dashboard/terraform/examples/dashboard"
    ```

### Preparing the Recommendations Dashboard Project

In this section you prepare your project for deployment.

1.  Go to the [project selector page][project-selector] in the Cloud Console.
    Select or create a Cloud project.

2.  Make sure that billing is enabled for your Google Cloud project.
    [Learn how to confirm billing is enabled for your project][enable-billing].

3.  In Cloud Shell, set environment variables with the ID of your **dashboard**
    project:

    ```sh
    export PROJECT_ID=<INSERT_YOUR_PROJECT_ID>
    gcloud config set project "${PROJECT_ID}"
    ```

4.  Choose the [region][region-and-zone] infrastructure will be located.

    ```sh
    export REGION=us-central1
    ```

5.  Enable the required Cloud APIs

    ```sh
    gcloud services enable cloudscheduler.googleapis.com \
      cloudresourcemanager.googleapis.com \
      bigquery.googleapis.com \
      workflows.googleapis.com \
      iam.googleapis.com
    ```

### Deploying the Dashboard

1. Change directory and set the necessary input variables either directly in the Terraform file `terraform.tfvars`. 

   The required variables are as follows:
   - `project_id`: The ID of the GCP project where the resources will be deployed.
   - `organization_id`: The ID of the GCP organization.
   - `region`: The region where the workflows and scheduler job will be created.

   ```sh
    cd "${DASHBOARD_DIR}"
    vi terraform.tfvars
    ```

2. Initialize the Terraform example dashboard.

    ```sh
    terraform init
    ```

3. Create the Dashboard infrastructure. Answer `yes` when prompted, after
    reviewing the resources that Terraform intends to create.

    ```sh
    terraform apply
    ```

4. Deploy the Looker Studio dashboard, by copying and pasting the url outputted by the Terraform into a browser. The report will load with no data since your exports have not yet been executed.

5. Click the "Edit and Share" button.

6. Review the data source updates and click "Acknowledge and save".

7. Click the "Add to Report" button.

8. Click the "View" button.

## Repository Structure

This repository contains the following folders:

- `resources`: Images and supporting materials for the readmes.
- `sql`: Contains the BigQuery views used to create the dashboards.
- `terraform`: Contains the Terraform code used to create the workflow script, Cloud Scheduler, and BigQuery tables.

[activeassist]: <https://cloud.google.com/active-assist>
[assetinventory]: <https://cloud.google.com/asset-inventory>
[bigquery]: <https://cloud.google.com/bigquery>
[lookerstudio]: <https://cloud.google.com/looker>
[project-selector]: https://console.cloud.google.com/projectselector2/home/dashboard
[enable-billing]: https://cloud.google.com/billing/docs/how-to/modify-project
[cloud-console]: https://console.cloud.google.com
[cloud-shell]: https://console.cloud.google.com/?cloudshell=true
[region-and-zone]: https://cloud.google.com/compute/docs/regions-zones#locations