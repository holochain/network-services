# network-services
A Pulumi definition for deploying Holochain network services to be used for development

## Development

### Installation

First, make sure that you are in the Nix development shell or that you have
installed `pulumi`, `pulumi-go`, and `go`.

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

## Bootstrap server

The bootstrap server is a combination of [bootstrap](https://crates.io/crates/kitsune2_bootstrap_srv) and 
[sbd (Holochain signal server)](https://crates.io/crates/sbd-server) services. Used for peer discovery and initiating peer connections 
respectively.

The services are deployed as a container which can be found in the [Kitsune2 project](https://github.com/holochain/kitsune2/pkgs/container/kitsune2_bootstrap_srv).

### Setting up on the first deploy

The first step is to create a deployment configuration. It includes the `docker-compose.yaml` that defines the services in the docker container,
and the `cloud-init.yaml` which triggers the deployment of the defined container. One of the existing folders `dev-test` or the like can be
copied and the file contents adapted to the configuration of the new deployment.

The new deployment needs to be included in `./main.go` to be executed by pulumi. Again, existing code can be copied and modified as needed.
Create a PR with the changes so far and merge it into `main`. The new deployment will be performed automatically.

The next step is to set up DNS. You need to map the intended hostname for the server to the IPv4 and IPv6 addresses of
the server.

Certificates are required to run the bootstrap/sbd service. Generating them with [certbot](https://certbot.eff.org/) is
an interactive process. You need to run the following command and follow the instructions:

```sh
sudo certbot certonly --standalone -d dev-test-bootstrap2.holochain.org
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

### Updating the container deployment

To update the container deployment, edit the `docker-compose.yaml` locally. Any changes to this file should go up for a
pull request. Once that's done, you need to run the following command locally:

```sh
scp dev-test/docker-compose.yaml root@dev-test-bootstrap2.holochain.org:/opt/bootstrap_srv/docker-compose.yaml
```

Then, you need to SSH into the server and restart the service:

```sh
ssh root@dev-test-bootstrap2.holochain.org
cd /opt/bootstrap_srv/
podman compose up -d
```

Note that this will restart the service, which will close any open connections!

## Bootstrap server with authentication

The same as the bootstrap server, but with authentication. This is used to protect the bootstrap and sbd servers from
abuse and is intended to become the default in the future.

### Setting up on the first deploy

The first step is to set up DNS. You need to map the intended hostname for the server to the IPv4 and IPv6 addresses of
the server.

Certificates are required to run the bootstrap/sbd service. Generating them with [certbot](https://certbot.eff.org/) is
an interactive process. You need to run the following command and follow the instructions:

```sh
sudo certbot certonly --standalone -d dev-test-bootstrap2-auth.holochain.org
```

This sets up auto-renewal of certificates, so there's no need to manually configure a cron job. You can check that this
is working by running:

```sh
sudo certbot renew --dry-run
```

Now you can start the bootstrap/sbd and authentication services:

```sh
cd /opt/bootstrap_srv/
podman compose up -d
```

This will pull the `holochain/kitsune2_bootstrap_srv` and `holochain/kitsune2_test_auth_hook_server` containers and start them as a daemon.

Check that the service managed to start by running:

```sh
podman compose logs bootstrap
```

You should see a log line like `#kitsune2_bootstrap_srv#running#`.

### Updating the container deployment

To update the container deployment, edit the `docker-compose.yaml` locally. Any changes to this file should go up for a
pull request. Once that's done, you need to run the following command locally:

```sh
scp dev-test-auth/docker-compose.yaml root@dev-test-bootstrap2-auth.holochain.org:/opt/bootstrap_srv/docker-compose.yaml
```

Then, you need to SSH into the server and restart the service:

```sh
ssh root@dev-test-bootstrap2-auth.holochain.org
cd /opt/bootstrap_srv/
podman compose up -d
```

Note that this will restart the service, which will close any open connections!

## Cloudflare DNS

DNS records for newer deployments are managed by Pulumi using the Cloudflare provider. This requires a Cloudflare API
token and the zone ID for the `holochain.org` DNS zone.

Set the zone ID:

```shell
pulumi config set dns:cloudflare-zone-id <holochain.org-zone-id>
```

Set the API token as a secret:

```shell
pulumi config set --secret cloudflare:apiToken <cloudflare-api-token>
```

The API token needs `Zone:DNS:Edit` permission for the `holochain.org` zone.

## Iroh bootstrap server (dev-test)

A standalone Iroh relay bootstrap server without authentication, intended for development testing. DNS is provisioned
automatically via Cloudflare, and TLS certificates are obtained via certbot on first boot.

The service is deployed as a Podman Quadlet container managed by systemd. On first boot, cloud-init provisions the TLS
certificate (retrying until DNS propagates) and starts the `bootstrap` systemd service.

### First deploy

No manual setup is required beyond the Pulumi configuration. The deployment creates:
- A DigitalOcean droplet
- Cloudflare A and AAAA records for `dev-test-bootstrap2-iroh.holochain.org`

Cloud-init handles certbot and service startup automatically.

### Managing the service

SSH into the server and use systemd to manage the bootstrap service:

```sh
ssh root@dev-test-bootstrap2-iroh.holochain.org

# Check service status
systemctl status bootstrap

# View logs
journalctl -u bootstrap

# Restart
systemctl restart bootstrap
```

### Updating the container

Edit `dev-test-bootstrap2-iroh/cloud-init.yaml` locally and open a PR. Once merged, the cloud-init change will take effect on the
next droplet recreation. To update a running instance, SSH in and update the Quadlet container file:

```sh
scp dev-test-bootstrap2-iroh/cloud-init.yaml root@dev-test-bootstrap2-iroh.holochain.org:/tmp/cloud-init.yaml
```

Then on the server, extract the updated container definition from the cloud-init and reload:

```sh
# Edit /etc/containers/systemd/bootstrap.container with the new image/config
systemctl daemon-reload
systemctl restart bootstrap
```

## Iroh-flavored bootstrap service with authentication for Unyt

Configured an OAuth application as directed by the [hc-auth-iroh-unyt README](https://github.com/holochain/hc-auth-server).

Set the homepage URL to: https://github.com/holochain/hc-auth-server
Set the callback URL to: https://hc-auth-iroh-unyt.holochain.org/ops/oauth-callback

Then, set the following Pulumi configuration values with the credentials from the OAuth application:

```shell
pulumi config set --secret hc-auth-iroh-unyt:github-client-id <value>
pulumi config set --secret hc-auth-iroh-unyt:github-client-secret <value>
```

Next, generate a session secret and set it as a secret:

```shell
dd if=/dev/random bs=66 count=1 | base64 -w0 | pulumi config set --secret hc-auth-iroh-unyt:session-secret
```

Finally, set an API token for the authentication server. This can be any random string, but it must be kept secret. Use
the output api-token.txt to share the token with users who need it and then remove it:

```shell
uuidgen | tee api-token.txt | pulumi config set --secret hc-auth-iroh-unyt:api-tokens
```
