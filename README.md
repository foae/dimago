# dimago
Created a diagram containing dependencies for any Go project hosted on Github.com platform.

### Getting up and running
* `go get -u github.com/foae/dimago`
* `make test && make run` – will start an HTTP server on localhost 8080

### Methods
#### GET `/`
```json
{
    "message": "OK",
    "status": 200
}
```
#### POST `/`
* Action: will retrieve the repository for inspection
* Result: a diagram served via Cacoo API that resembles [this example](https://i.imgur.com/FBOZBpD.png) – _**under development!**_
* Needed headers: `Content-Type: application/json`
* Body:
```json
{
    "url": "https://github.com/username/projectname"
}
```