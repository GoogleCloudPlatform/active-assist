# Google Cloud Org Level Recommendations Hub

Google Cloud's [Active Assist][activeassist] tooling creates recommendations to improve the operation of your Google Cloud resources, including cost efficiency, security, and sustainability. Most recommendations are scoped and only visible at the project level. This tooling sets up exports for Recommendations and [Cloud Asset Inventory][assetinventory] to [BigQuery][bigquery] and provides [Looker Studio][lookerstudio] dashboards. The dashboards create a view into the recommendations across all the projects in your organization, with an emphasis on highlighting opportunities for cost savings.

## Getting Started

### Prerequisites

1. Google Cloud Organization
2. Google Cloud Support Purchased for your Organization.

## Repository Structure

This repository contains the following folders:

- `sql`: Contains the BigQuery views used to create the dashboards.
- `terraform`: Contains the Terraform code used to create the workflow script, Cloud Scheduler, and BigQuery tables.

## Usage

Please refer to the respective folders (`sql` and `terraform`) for detailed instructions on setting up and configuring the dashboards and infrastructure components.

[activeassist]: <https://cloud.google.com/active-assist>
[assetinventory]: <https://cloud.google.com/asset-inventory>
[bigquery]: <https://cloud.google.com/bigquery>
[lookerstudio]: <https://cloud.google.com/looker>
