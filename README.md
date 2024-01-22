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
$ curl 'localhost:5000/ping'
{ "status": "pong" }
```

## Usage

### View Recent Repositories

##### Basic usage

This endpoint will return the most recent 100 repositories on GitHub when called without any filters

```
$ curl 'localhost:5000/repos'
[
 {"full_name":"ownerName/repoName","owner":"ownerName","repository":"repoName","languages":{"html":{"bytes":564},"javascript":{"bytes":6469},"scss":{"bytes":3074}}},
...
]
```

##### Search using filters

Additionally you can filter by language. This will limit the number of results to only those which are known to use the language specified in the query parameter.

It will return info about repos that contain more than one language as long as at least one of them matches a language specified in the filter. You can specify more than one language to search for by concatenating with a comma.

```
$ curl 'localhost:5000/repos?language=ruby,python'
[
 {"full_name":"ownerName/repoName","owner":"ownerName","repository":"repoName","languages":{"python":{"bytes":4932}}},
 {"full_name":"ownerName/repoName","owner":"ownerName","repository":"repoName","languages":{"html":{"bytes":564},"ruby":{"bytes":6469},"scss":{"bytes":3074}}},
...
]
```

## Configuration

Here are some environment variables that can be used to tweak the application performance

#### WORKER_COUNT

number of workers (per search request) which can make concurrent requests to fetch language data for repos

#### AUTH_INTERVAL

a duration which marks how often the application attempts to refresh the auth token

## Design Considerations

The project is divided into 4 main components:

* Github Client - for isolating business logic related to GitHub's API and managing API requests
* Authenticator - for managing the authentication lifecycle and refresh of GitHub API tokens
* Searcher - for discovering repositories relevant to the search
* Subrequester - for making subsequent requests concurrently via multiple workers

#### Approach

The application uses the [list public events API](https://docs.github.com/en/rest/activity/events?apiVersion=2022-11-28#list-public-events) to first search for the latest repository_id (n) and then use that number to calculate the n-100th repository.

It is then able to pass that repository_id as the 'since' value to the [list public repositories API](https://docs.github.com/en/rest/repos/repos?apiVersion=2022-11-28#list-public-repositories). This approach was taken because the list public repositories API does not expose a sorting mechanism which would allow repositories to be fetched in descending order.

A caveat to this is that the list public repos API response does not include all the details about a repository (like licence information for example), so we are not able to extract as many details that might be interesting to filter by.
However the advantage of being able to get all 100 repositories in one HTTP call is significant in terms of both speed and also avoiding maxing out the rate-limit.

The rate-limit in particular is a significant limitation in terms of the scaling of this application. Although using proper authentication methods does increase the rate-limit, we have to keep in mind that we are only granted 5K req/h (see: [docs on rate limits](https://docs.github.com/en/rest/using-the-rest-api/rate-limits-for-the-rest-api?apiVersion=2022-11-28)) so horizontally scaling of this application would quickly cause it to be throttled.

### Next steps

#### Caching

Caching would be an obvious next step towards optimising this application. Currently all requests (with exception of authentication which has its own lifecycle) are being triggered via the /repos handler and no effort to reuse results is made.

In order to determine if caching would be effective we'd first have to better define the criteria for 'latest repos' and also better understand the typical use patterns of the app.
Since git repositories are being created on the fly all the time, two subsequent requests to the /repos endpoint are unlikely to yield the same result set. If the emphasis is really on having the latest repos then caching is unlikely to have a huge impact as most of the data would already be stale by the time it is saved in the cache.

However, if for example the users are not literally interested only in the most recent 100 repositories, but more generally in repositories that have been created in the last hour or so, we could potentially create a cache that stores historical data.
This could expand on the usefulness of the ?languages filter, which instead of returning only those entries of the most recent 100 which match the language, could instead return the last 100 repositories matching that criteria within a particular time frame.

#### Graceful Shutdown

The docker-compose configuration was initially configured to send a SIGKILL to the application, however with web servers it is a typical practice to attempt to wait for underlying goroutines to avoid abruptly hanging up on HTTP clients still connected to the server.

The handling surrounding ListenAndServe has been modified to allow the catching of SIGTERM / SIGINT and attempt to close the webserver.

The SIGKILL directive in the docker-compose configuration has been removed so that it will now default to sending and initial SIGTERM directive to give the application a chance to shutdown gracefully.

### Regarding filters

I only implemented a single ?languages filter for the following reasons:
* licencing information was not available in the response of the [list public repositories API](https://docs.github.com/en/rest/repos/repos?apiVersion=2022-11-28#list-public-repositories)
* language was the only attribute visible in the requested API response
* it was unclear from the task description which other data points (stargazers count, pull request count?) would be interesting to filter by
* the description of the task put more emphasis on speed of response and scalability

So, as mentioned above, I prioritised reducing the total number of requests to avoid being easily throttled by the API rate limit over adding additional filters.

In a real world situation it would of course be important to fully analyse the user needs (eg. why the user is searching for repos and what kind of information they expect to find) before deciding on the final functionality / tradeoffs to be made.
