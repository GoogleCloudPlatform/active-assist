# Recommender Data Export

This Google Workflows YAML file is used to export recommendation data from the Recommender API to BigQuery. It allows you to specify the organization, dataset, and table where the exported data will be stored. 

## Prerequisites

- The script requires the following Google Cloud APIs to be enabled:
  - Cloud Scheduler API
  - BigQuery API
  - Workflow Executions API
  - Workflows API

## Usage

The script is written in YAML format and should be deployed and executed as a Google Cloud Workflow.

### Workflow Variables

The following variables need to be set in the YAML file before deploying and executing the workflow:

- `orgId`: The organization ID you want to export.

### Optional Variables

The following variables can be set in the YAML file to customize the export process:

- `datasetId`: The BigQuery dataset where the export will take place. If not provided, the default value is "recommendations_export_dataset".
- `assetTable`: The table in BigQuery where the exported data will be stored. If not provided, the default value is "asset_export_table".
- `levelToExport`: The Org/Project/Folder level at which the exports will occur. For example, "organizations/123" or "projects/my-project-id". Note that exporting at the project level may require calling the API directly. If not provided, the default value is "organizations/{orgId}".
- `bqLocation`: The location of the BigQuery dataset. If not provided, the default value is "US".
- `projectId`: The project ID for the BigQuery dataset. If not provided, the default value is the `GOOGLE_CLOUD_PROJECT_ID` environment variable.

### Deploying and Executing the Workflow

1. Ensure that the prerequisites are met and the required APIs are enabled in your Google Cloud project.
2. Set the desired values for the variables in the YAML file.
3. Deploy the workflow by running the following command:

```bash
gcloud workflows deploy recommender-data-export --source=[PATH_TO_YAML_FILE]
```


4. Once the workflow is deployed, you can manually execute it by running the following command:

```bash
gcloud workflows execute recommender-data-export
```

The workflow will start executing and export the recommendation data to the specified BigQuery dataset and table.

## Permissions

The service account executing the workflow requires the following permissions:

- **Cloud Asset Permissions**:
  - cloudasset.assets.exportResource

- **BigQuery Permissions**:
  - bigquery.datasets.create
  - bigquery.tables.create
  - bigquery.tables.delete
  - bigquery.tables.export
  - bigquery.dataViewer

- **Service Usage Permissions**:
  - serviceusage.services.enable

## Clean Up

To remove the exported data and associated resources, you can manually delete the BigQuery dataset and table that were created.
