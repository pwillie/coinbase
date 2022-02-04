## Technical Assessment (SRE)

Part 1) Using the language of your choice, create a simple webservice that returns a json object containing spot price data from coinbase (https://api.coinbase.com/v2/prices/spot?currency=USD) on requests to `/<currency>`.

The endpoint should minimally support: EUR, GBP, USD and JPY. This service should be run in a container. The container should be built via github actions.

Include a `/health` endpoint that returns 200 if the application is running.


Part 2) Expand the Github Actions used to create the container to create some simple tests. The test should attempt to connect to the mock service and should fail if the json object cannot parsed, or if the currency does not exist.


Optional extras:
  - Include a `/metrics` or `/health` endpoint that reports simple health metrics
  - Integrate slack messaging into the github actions or directly into the application
  - Build a helm chart to deploy the service via github actions
  - Write terraform that can deploy the container into AWS (ECS, EKS, ec2, lambda)