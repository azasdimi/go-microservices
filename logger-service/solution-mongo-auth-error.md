If error is

connection() error occurred during connection handshake: auth error: sasl conversation error: unable to authenticate using mechanism "SCRAM-SHA-1": (AuthenticationFailed) Authentication failed.

Then go to terminal

Get list of containers using podman ps -a

Get mongo container id

Then run

podman exec -it <container-id> mongo
use admin
db.createUser({
user: "admin",
pwd: "password",
roles: [{ role: "readWrite", db: "logs" }]
})
