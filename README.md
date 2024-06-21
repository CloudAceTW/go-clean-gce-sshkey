# Cleaning Up GCE Metadata: Removing Unused SSH Keys

## Before Use
This script requires a Golang runtime environment to be set up on your machine.

## How to Use
1. There are two ways to authenticate with Google Cloud Platform (GCP) for this script:

- **gcloud CLI**: If you have already set up the gcloud command-line tool and authenticated your account, you can proceed without further steps.

- **Service Account**: You can authenticate using a service account by setting the GOOGLE_APPLICATION_CREDENTIALS environment variable. This variable should point to the path of the JSON key file for your service account. An example is provided below:
  ```
  export GOOGLE_APPLICATION_CREDENTIALS=<PATH/FOR/SERVICE_ACCOUNT/JSON>
  ```
2. This script helps you remove unused SSH keys from the metadata of Compute Engine instances within a GCP project.
```
go run ./main.go -projectID=gcp-project-id -users=user1,user2,user3
```
- projectID: (Required) The ID of your GCP project.
- users: (Required) Comma-separated list of usernames for which you want to remove unused SSH keys.

