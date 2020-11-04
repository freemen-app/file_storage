# File Storage
gRPC API that allows yuo to upload files to object storage (Amazon S3)

## Env
The following variables have to be set to `.env` file to run this app using docker.
* AWS_BUCKET
* AWS_ACCESS_KEY_ID
* AWS_SECRET_ACCESS_KEY
* AMQP_USERNAME
* AMQP_PASSWORD
* AMQP_HOST (optional default: localhost)
* AMQP_PORT (optional, default: 5672)

## Running
```
docker-compose up
```
