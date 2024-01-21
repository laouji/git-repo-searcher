# git-repo-searcher

## Prerequisites

* You must have created a [GitHub App](https://docs.github.com/en/apps/creating-github-apps)
* You must have generated a private key for this app and have it on your local machine
* You must have clicked ["Install App"](https://docs.github.com/en/apps/using-github-apps/installing-your-own-github-app) on GitHub and associated with your account

* docker compose must be available on your local machine

## Execution

The most basic usage is to run the web server using the up command:
```
$ make up
```

While you can run this application without setting any environment variables, you will quickly notice that is easy to hit the GitHub API rate-limit unless you pass github authentication credentials into the app and allow it to authenticate.

1. Find the client ID for your Github App at the top of the App Settings page in Github

(Settings -> Developer settings  -> GitHub Apps -> YOUR APP)

2. From the same GitHub page generate and download a private key with the .pem extension

3. From the home directory of the project execute the following:
```
$ export GITHUB_APP_ID=${YOUR-CLIENT-ID} && cat ${PATH-TO-YOUR-GITHUB-API-PRIVATE-KEY} > ./private_key.pem && make up
```

Application will be then running on port `5000`

## Test

```
$ curl localhost:5000/ping
{ "status": "pong" }
```

## Usage

### View Recent Repositories

##### Basic usage

This endpoint will return the most recent 100 repositories on GitHub when called without any filters

```
$ curl localhost:5000/repos
[
 {"full_name":"ownerName/repoName","owner":"ownerName","repository":"repoName","languages":{"html":{"bytes":564},"javascript":{"bytes":6469},"scss":{"bytes":3074}}},
...
]
```

##### Search using filters

Additionally you can filter by language. This will limit the number of results to only those which are known to use the language specified in the query parameter.

It will return info about repos that contain more than one language as long as at least one of them matches a language specified in the filter. You can specify more than one language to search for by concatenating with a comma.

```
$ curl localhost:5000/repos?language=ruby,python
[
 {"full_name":"ownerName/repoName","owner":"ownerName","repository":"repoName","languages":{"python":{"bytes":4932}}},
 {"full_name":"ownerName/repoName","owner":"ownerName","repository":"repoName","languages":{"html":{"bytes":564},"ruby":{"bytes":6469},"scss":{"bytes":3074}}},
...
]
```
