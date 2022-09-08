## Prepare your host project

1. Required IAM roles

Remora will require the creation of some organization level resources. The user or service 
account that will run the Terraform will need the following IAM roles on the organization IAM policy:
- Organization Role Administrator
- Organization Administrator

Remora will require a host project to create its resources. The user or service account
that will run the Terraform will need the following IAM roles on the project IAM policy:

- Secret Manager Admin
- Service Usage Admin
- Create Service Accounts
- Delete Service Accounts
- Pub/Sub Editor
- Storage Admin
- Workflows Editor
- Service Account User
- Cloud Functions Developer
- Cloud Scheduler Admin
- Project IAM Admin

2. Upload Sendgrid or Jira Key to Google Cloud Secret Manager

   Once you have this project and permissions, you will need to add your Sendgrid or Jira key to Secret Manager, depending on what solution
   will be used to notify project owners of pending project deletion:

- If using Sendgrid, you will use Secret Manager to manage the Sendgrid API key. You will need this key name (_sendgrid-key_) in the Terraform setup as the `sendgrid_secret_name` variable:

```
gcloud secrets create sendgrid-key \
    --replication-policy="automatic"
gcloud secrets versions add sendgrid-key --data-file=${SENDGRID_API_KEY_PATH}
```

- If using Jira, you will use Secret Manager to manage the Jira API key. You will need this key name (_jira-key_) in the Terraform setup as the `jira_secret_name` variable:

```
gcloud secrets create jira-key \
    --replication-policy="automatic"
gcloud secrets versions add jira-key --data-file=${JIRA_API_KEY_PATH}
```

## Prepare and run Terraform configuration

1. Starting in one of the example directories, populate `terraform.tfvars.example` with the required values. See `variables.tf` for all variables, their types and descriptions.
1. Rename `terraform.tfvars.example` to `terraform.tfvars`
1. Initialize terraform: `terraform init`
1. Preview the resources Terraform will create: `terraform plan`
1. Apply the Terraform configuration: `terraform apply`
