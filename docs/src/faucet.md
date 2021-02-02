# Token Faucet

For the main agent land network we have a simple token faucet implemented to allow users of the network to get started quickly.

Token Faucets are network specific, depending on the network type they may or may not be deployed. Please check the [networks](../networks/) page for specific details.

The Token Faucet itself is a HTTP REST API and interaction is shown below:

## Making a token claim request

A user must first submit a token claim request to the faucet. This request can be either URL encoded or JSON encoded. Both examples are shown below:

URL: `POST /claim/requests`

### URL Encoded

**Headers**

The request must include the following headers:

```
'Content-Type: application/x-www-form-urlencoded
```

**cURL example**

```bash
curl -d 'Address=fetch1xqqftqp8ranv2taxsx8h594xprfw3qxl7j3ra2' -H "Content-Type: application/x-www-form-urlencoded" -X POST http://127.0.0.1:5000/claim/requests
```

### JSON Encoded

**Headers**

The request must include the following headers:

```
'Content-Type: application/json
```

**cURL example**

```bash
curl -d '{"Address":"fetch1xqqftqp8ranv2taxsx8h594xprfw3qxl7j3ra2"}' -H "Content-Type: application/json" -X POST http://127.0.0.1:5000/claim/requests
```

### Response

In either submission case, upon successful register the API will respond with the request UID

```json
{
  "uid": "123e4567-e89b-12d3-a456-426614174000"
}
```

## Querying the status of a token claim request

To query the status of the your token claim request the following API can be used

URL: `GET /claim/requests/<uid>`

**cURL example**

```bash
curl http://127.0.0.1:5000/claim/requests/1a472a27-7225-4409-92be-4efd3beed995
```

This will respond with the status of the request claim. If the claim was successful then JSON response will resemble the following:

```json
{
  "txDigest": "9CA3A7D3614A37C1BB2EA6B746B402CF68D3E5A4CEBFFE1D7ADF212876DAE70B",
  "status": "completed",
  "statusCode": 20,
  "lastUpdated": "2020-08-11T09:48:04.596522"
}
```

## Rate limiting

To prevent malicious actors, this API is rate limited and will block requests if the limits are passed. In this case the user must wait and then try again.
