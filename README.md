# Payment service

Implement a new microservice for handling SCB internet banking payment channels
The service connects to omise payment  gateway and completes customer fulfillment. 
The service can: 
- create a new payment request 
- record transactions status 
- query for a transaction status 

## Deployment steps - On Docker
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

Example for request payloads

```json
{
  "object": "event",
  "id": "evnt_test_xxxxxxxx",
  "livemode": false,
  "location": "/events/evnt_test_xxxxxxxx",
  "webhook_deliveries": [
    {
      "object": "webhook_delivery",
      "id": "whdl_test_xxxxxxxx",
      "uri": "https://omise-flask-example.herokuapp.com/webhook",
      "status": 200
    }
  ],
  "data": {
    "object": "charge",
    "id": "chrg_test_xxxxxxxx",
    "location": "/charges/chrg_test_xxxxxxxx",
    "amount": 12345,
    "net": 11862,
    "fee": 451,
    "fee_vat": 32,
    "interest": 0,
    "interest_vat": 0,
    "funding_amount": 12345,
    "refunded_amount": 0,
    "authorized": true,
    "capturable": false,
    "capture": true,
    "disputable": true,
    "livemode": false,
    "refundable": true,
    "reversed": false,
    "reversible": false,
    "voided": false,
    "paid": true,
    "expired": false,
    "platform_fee": {
      "fixed": null,
      "amount": null,
      "percentage": null
    },
    "currency": "THB",
    "funding_currency": "THB",
    "ip": "203.0.113.1",
    "refunds": {
      "object": "list",
      "data": [],
      "limit": 20,
      "offset": 0,
      "total": 0,
      "location": "/charges/chrg_test_xxxxxxxx/refunds",
      "order": "chronological",
      "from": "1970-01-01T00:00:00Z",
      "to": "2019-12-31T12:59:59Z"
    },
    "link": null,
    "description": null,
    "metadata": {
      "order_id": "P26042018-01",
      "color": "pink"
    },
    "card": {
      "object": "card",
      "id": "card_test_xxxxxxxx",
      "livemode": false,
      "location": null,
      "deleted": false,
      "street1": "1448/4 Praditmanutham Road",
      "street2": null,
      "city": "Bangkok",
      "state": null,
      "phone_number": "0123456789",
      "postal_code": "10320",
      "country": "th",
      "financing": "credit",
      "bank": "Bank of the Unbanked",
      "brand": "Visa",
      "fingerprint": "XjOdjaoHRvUGRfmZacMPcJtm0U3SEIIfkA7534dQeVw=",
      "first_digits": null,
      "last_digits": "4242",
      "name": "Somchai Prasert",
      "expiration_month": 12,
      "expiration_year": 2022,
      "security_code_check": true,
      "created_at": "2019-12-31T12:59:59Z"
    },
    "source": null,
    "schedule": null,
    "customer": null,
    "dispute": null,
    "transaction": "trxn_test_xxxxxxxx",
    "failure_code": null,
    "failure_message": null,
    "status": "successful",
    "authorize_uri": "https://api.omise.co/payments/paym_test_xxxxxxxx/authorize",
    "return_uri": "https://www.example.com/orders/54321/complete",
    "created_at": "2019-12-31T12:59:59Z",
    "paid_at": "2019-12-31T12:59:59Z",
    "expires_at": "2019-12-31T12:59:59Z",
    "expired_at": null,
    "reversed_at": null,
    "zero_interest_installments": true,
    "branch": null,
    "terminal": null,
    "device": null
  },
  "key": "charge.create",
  "created_at": "2019-12-31T12:59:59Z"
}
```