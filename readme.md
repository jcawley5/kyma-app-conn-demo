## Kyma Application Connection Demo

Demostrates the application connection process in a step by step manner. The example supports both the graphql api provided by the compass management plane as well as the rest api provided by the kyma application connector.

### Instructions
- Use the deployment.yaml to deploy the app to kyma.  
- This will generate an API to access the app.
- Provide either a management plane token or a kyma applicaton connector token and use the `Call Token URL` to initialize the process.
- Process each of the following steps in the order shown.
  
### API
- An example api exists at `/orders`.  Each event triggered will populate corresponding data in the api.


