# Ethereum Blockchain Parser

## Goal

Implement Ethereum blockchain parser that will allow to query transactions for subscribed
addresses.

## Problem

Users not able to receive push notifications for incoming/outgoing transactions. By
Implementing Parser interface we would be able to hook this up to notifications service to
notify about any incoming/outgoing transactions.

## Reference

- [eth_blockNumber](https://ethereum.org/en/developers/docs/apis/json-rpc/#eth_getblockbynumber)
- [eth_getblockbynumber](https://ethereum.org/en/developers/docs/apis/json-rpc/#eth_getblockbynumber)

## Howto

### Test

```bash
make test
```

### Build

```bash
make build
```

### Lint

```bash
make lint
```


### Run app

```bash
make run
```

We can pickup an address from the logs and query the transactions for that address.

![use the http api](./httpapi.gif)

## TODO

- [x] Use Mockery to mock the api interface and test the processBlock method in the `parser.go`
- [ ] Use a limiter to limit the number of concurrent requests to the Ethereum API to avoid 429 errors
- [ ] Extract the memory store to a separate package and can be replaced with a persistent store in the future