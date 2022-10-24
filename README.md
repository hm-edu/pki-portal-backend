# PKI-Portal Backend

This project contains all backend services required for running a self-service PKI-Protal providing a middleware between GEANT-TCS sectigo and the endusers. It is a mixture out of enduser facing REST-APIs and some internal GRPC service communicating with sectigo and other services like the DNS backend.

![](docs/overview.png)

## Prerequirements

At the moment the setup is a little complex and requires some manual steps and interaction. Feel free to provide some optimizations and improvments. In the best case you run the complete setup in an existing k8s cluster and adapt the deployment located [here](https://github.com/hm-edu/infrastructure/tree/main/clusters/production/portal). Otherwise, you can run the setup using a local docker engine or even running single exectuables on bare metal.

### Docker Deployment

The docker deployment is the easiest way to get started. It requires a running docker engine and docker-compose. The docker-compose file is located inside this repository. 

Before getting started you must configure at least your own sectigo credentials and set the organization ID for ssl and smime certificates.

```bash
cp .env.example .env
vim .env
```

Afterwards you can start the whole setup using the following command:

```bash
docker-compose up -d
```

For a more sophisticated setup you can of cource configure the docker-compose file to your needs and define other values and parameters.
