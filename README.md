The purpose of this simple application is mocking a simple HTTP server with a YAML file.
In many cases such as developing front-end or client application, we want to mock the backend or api server.

With this application, you can easily mock the result of the HTTP requests based on some conditions if needed.

# How to use this application

Firstly, you need to prepare the configuration file in YAML format.

## Without docker
After cloning the repository, build this application with
```bash
go build
./yamlserver --config <path_to_config.yaml>
```

## With docker
After cloning the repository, run:
```bash
docker build -t yamlserver .
docker run -v `pwd`/example/config.yaml:/config.yaml -p 8062:8062 yamlserver
# a server will be run at port 8062
```

This is an example of config file:
```yaml
server:
  host:
  port: 8062

endpoints:
  - path: /ping
    method: # any method
    results:
      - when:
        response:
          returnObject:
            message: "pong"
  - path: /
    method: "GET"
    results:
      - when:
          query:
            time: 10
            age: 30
        response:
          returnCode: 200
          returnObject:
            - "item 1"
            - "item 2"
      - response: # no condition == default return
          returnCode: 200
          returnObject:
            data1: this is data1
            data2: this is data2
      - when:  # Same as above, but it's nerver reach here because the above item matched
        response: # no condition == default return
          returnCode: 200
          returnObject:
            data1: this is data1
            data2: this is data2
  - path: /
    method: "POST"
    results:
      - when:
          payload:
            password: "a too simple password"
            address:
              primary: "France"
        response:
          returnCode: 400
          returnObject: BAD REQUEST
      - when: # in any case
        response:
          returnCode: 500
          returnObject: 
```

With this configuration, we define three endpoints:
- `GET /`: There are two different outputs, depends on the request
  * 1/ If the URL query `time=10&age=30`, return a list `[ "item 1", "item 2"]`
  * 2/ The default response, with no constraint in the request
- `POST /`: There are two different outputs, depends on the request
  * 1/ If the payload contains the `password` `a too simple password`, and the `address.primary` is `France`, we return `BAD REQUEST` with status code `400`.
  * 2/ By default, return status code `500`
- `GET /ping`: A pong message is returned

Currently, we support to set the constraints on `header`, `query` and `payload`.
With a given HTTP request, the application scans through **all** endpoints, compare the request information to the constraint of each endpoint and each scenario to find the **first** match.


# To build image
```bash
docker build -t ductrungnguyen/yamlserver:`cat version` .
```

# To push image to Docker hub
```bash
docker push ductrungnguyen/yamlserver:`cat version` 
```