# Payment service

Implement a new microservice for handling SCB internet banking payment channels
The service connects to omise payment  gateway and completes customer fulfillment. 
The service can: 
- create a new payment request 
- record transactions status 
- query for a transaction status 

## Prerequisites
- Receive Public key and Secret key at https://dashboard.omise.co/test/keys
then update in docker-compose.yml ( OMISE_PUBLIC_KEY, OMISE_SECRET_KEY )
- Run command 
```sh
docker-compose up -d
```
- Get the webhook endpoint by executing the command line below

for general Docker
```sh
echo "https://$(curl --silent $(docker port payment_webhook_server_ngrok_tunnel 4040)/api/tunnels | sed -nE 's/.*public_url":"https:..([^"]*).*/\1/p')/webhook/omise"
```

for Docker on host
```sh
echo "https://$(curl --silent http://<DOCKER_HOST_IP>:4040/api/tunnels | sed -nE 's/.*public_url":"https:..([^"]*).*/\1/p')/webhook/omise"
```

Then update your webhook endpoint on https://dashboard.omise.co/test/webhooks


## API Specs
- Create payment

```
POST /payments
```

Example for request payloads
```json
{
    "amount": 2000,
    "currency": "thb",
    "returnUri": "https://example.com",
    "sourceType": "internet_banking_scb"
}
```

Example for response payloads
```json
{
    "chargeId": "chrg_test_xxxxxxxxx",
    "sourceId": "src_test_xxxxxxxxx",
    "authorizeUri": "https://pay.omise.co/offsites/ofsp_test_xxxxxxxxx/pay"
}
```

- Get payment status
```
GET /payments/charges/:chargeID/status
```
Example for response payloads
```json
{
    "status": "successful"
}
```

- Webhook from Omise service
```
POST /webhook/omise
```