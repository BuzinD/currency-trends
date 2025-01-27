# Currency-trends-monitor
**Whats this?** This is a project for trying use go lang futures.
## Start service for development:
- CLone env files
    ___be careful check it can rewrite your *.env in ./app/env/ skip this step if alrady has *.env files___ 
    </br> run `make cloneEnv` 
- follow to ./app/env dir and set the real values for variables in *.env files
- run: `make start-db` for start pg-db in container

## Run service
    - prepare your env files
    - prepare your docker image
    - run `make build-app`
    - run `make start` for starting services

## Fatures of project:
- Getting currency values
- 