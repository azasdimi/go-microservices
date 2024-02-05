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
