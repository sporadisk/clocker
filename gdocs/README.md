# Google Docs API interface

This feature reads and writes to a google cloud document instead of a local file.

## Requirements

- A google account
- [A google cloud project](https://developers.google.com/workspace/guides/create-project)
- That you've [enabled the Google Docs API](https://console.cloud.google.com/flows/enableapi?apiid=docs.googleapis.com) for your project
- One of:
  - Application-default credentials aquired via `gcloud auth application-default login`, if you've installed the gcloud CLI
  - Or a set of Oauth2 Desktop App credentials, [as described in the go quickstart guide](https://developers.google.com/docs/api/quickstart/go#authorize_credentials_for_a_desktop_application) - stored as `credentials.json` in the working directory from which you execute the command
