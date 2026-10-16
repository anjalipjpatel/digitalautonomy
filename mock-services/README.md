The payment mock runs the Mockoon environment imported from the OpenBanking API specification. To regenerate its Dockerfile, use the Mockoon environment file (not the OpenAPI specification):

```
mockoon-cli dockerize --data ./mock-services/payment/payment-mockoon.json --port 3000 --output ./mock-services/payment/Dockerfile
```

The account and events Dockerfiles were generated with these commands:

```
mockoon-cli dockerize --data ./mock-services/account/account-mockoon.json --port 3000 --output ./mock-services/account/Dockerfile
```

```
mockoon-cli dockerize --data ./mock-services/events/event-mockoon.json --port 3000 --output ./mock-services/events/Dockerfile
```