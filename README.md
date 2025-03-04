# network-services
A Pulumi definition for deploying Holochain network services to be used for development

## Development

### Installation

First, make sure that you are in the Nix development shell or that you have
installed `pulumi`, `pulumi-language-go`, and `go`.

Then, log into Pulumi with:
```sh
pulumi login
```

Next, set the default organisation to `holochain` with:
```sh
pulumi org set-default holochain
```

Finally, select the Pulumi stack that you want to use. For this repo it is `network-services`.
```sh
pulumi stack select network-services
```

### Making changes

Then preview the changes with:
```sh
pulumi preview
```

### Applying changes

Simply open a PR to see the preview of the changes in the CI. Then, once the PR
is reviewed and merged into the `main` branch, a new workflow will push the
changes to Pulumi.

## Changing the DigitalOcean token

Pulumi requires a Personal Access Token (PAT) for DigitalOcean to make calls to the API.

Currently, the PAT is linked to the `ThetaSinner` DigitalOcean user account. To
change the token, run the following command:
```sh
pulumi config set --secret digitalocean:token <new-token>
```

This value is encrypted by Pulumi and stored in [Pulumi.network-services.yaml].

Remember to open a PR with the new token and allow the CI/Actions to apply the
changes to Pulumi.

## Setting up on the first deploy

The first step is to set up DNS. You need to map the intended hostname for the server to the IPv4 and IPv6 addresses of
the server.

Certificates are required to run the bootstrap/sbd service. Generating them with [certbot](https://certbot.eff.org/) is
an interactive process. You need to run the following command and follow the instructions:

```sh
sudo certbot certonly --standalone -d bootstrap2.holochain.org
```

This sets up auto-renewal of certificates, so there's no need to manually configure a cron job. You can check that this 
is working by running:

```sh
sudo certbot renew --dry-run
```

Now you can start the bootstrap/sbd service:

```sh
cd /opt/bootstrap_srv/
podman compose up -d
```

This will pull the `holochain/kitsune2_bootstrap_srv` container and start it as a daemon.

Check that the service managed to start by running:

```sh
podman compose logs bootstrap
```

You should see a log line like `#kitsune2_bootstrap_srv#running#`.

## Updating the container deployment

To update the container deployment, edit the `docker-compose.yaml` locally. Any changes to this file should go up for a
pull request. Once that's done, you need to run the following command locally:

```sh
scp docker-compose.yaml root@bootstrap2.holochain.org:/opt/bootstrap_srv/docker-compose.yaml
```

Then, you need to SSH into the server and restart the service:

```sh
ssh root@bootstrap2.holochain.org
cd /opt/bootstrap_srv/
podman compose up -d
```

Note that this will restart the service, which will close any open connections!


