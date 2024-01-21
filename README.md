# Canvas for Backend Technical Test at Scalingo

## Prerequisites

* You must have created a GitHub App
* You must have generated a private key for this app and have it on your local machine
* You must have clicked ["Install App"](https://docs.github.com/en/apps/using-github-apps/installing-your-own-github-app) on GitHub and associated with your account

## Execution

```
export GITHUB_CLIENT_ID=${YOUR-CLIENT-ID}
cat ${PATH-TO-YOUR-GITHUB-API-PRIVATE-KEY} > ./private_key.pem && make up
```

Application will be then running on port `5000`

## Test

```
$ curl localhost:5000/ping
{ "status": "pong" }
```
