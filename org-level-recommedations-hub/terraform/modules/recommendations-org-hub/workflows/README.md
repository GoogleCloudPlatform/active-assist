# GCP Org Level Recommendations Exporter

A GCP Workflows script to setup and regularly export Asset Inventory and Recommender API suggestions to BigQuery

## Getting Started

### Prerequisites

1. GCP Organization
2. GCP Project with [Workflows](https://cloud.google.com/workflows) enabled where you would like to host this workflow. 
3. GCP Support Purchased for your Organization. 

### Setup

*As it currently stands, I've configured my environment to use the default service account and have not determined the actual permissions required. I am going to write this setup with my current set of instructions, not the end state.*

1. Find your default service account
2. Configure the default service account to use the following **ORG level** permissions:
    * Recommenders Exporter
    * Cloud Asset Viewer
3. Configure the default service account to use the following **Project Level** permissions:
    * Project Editor
    * BigQuery Admin
    * Service Usage Admin

### Workflow Deployment

1. Load up the GCP Console (console.cloud.google.com) and select the project you want to use.
2. Find GCP Workflows
3. Create a New Workflow, choosing the default service account for your service account. If you don't have the compute default service account, go and enable compute engine. 
4. Add the source code from [recommender-api-export-workflow.yaml](recommender-api-export-workflow.yaml).
5. Click Deploy and wait for the workflow to finish deploying. 

### Running the workflow

Currently as things stand, you will need to run the workflow manually entering in some information. To execute the workflow, click "Execute" and add the following input:

TODO(ghaun): Potentially we don't need all of these inputs. If nothing else they should become optional.

```
{
    "assetTable": "tableName", // Table name you would like for the asset inventory table.
    "bqLocation": "US", // This needs to be US for now as other regions are not supported
    "datasetId": "datasetName", // The name of the dataset you would like created
    "levelToExport": "organizations/123456789", // The Org ID you are in
    "orgId": "123456789", // The Org ID you are in
    "projectId": "project_id", // The project you are running this in
    "recommendationTable": "RectableName" // The name of the recommendations table you would like to use. 
}
```

### Other information.

The way this script operates is it will create the BQ Datasets and Tables for you. You do not need to create anything in BQ ahead of time. 

You also do not need to enable the apis in the console as this script should enable everything.

If you don't have any recommendations setup the workflow may not terminate and will continue to run until a transfer has been completed.