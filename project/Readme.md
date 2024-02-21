## Run the project in Podman

Start the podman VM

`podman machine init`
`podman machine start`

Navigate to the project folder

`cd project`

Start the containers

`make up_build`

Start the frontend app

`make start`

Stop the frontend app

`make stop`

Stop docker compose

`make down`

If PostgresSQL is not starting and shows address :5432 is already in use,
then

`sudo lsof -i :5432 `

Get the PID of the process and kill it

`sudo kill -9 <pid>`

## Building and pushing Images to Docker Hub using Podman

**_ Start podman if not started already _**

1. We need to login to podman

   `podman login -u <docker-username> -p <docker-password> docker.io/<username>`

2. Build the images

   ```
   cd /<folder>
   podman  build -f <application-name>.dockerfile -t azasdimi/<application-name>:1.0.0 .
   podman push azasdimi/<application-name>:1.0.0
   ```
