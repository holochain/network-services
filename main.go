package main

import (
	"bytes"
	"log"
	"os"
	"text/template"

	"github.com/pulumi/pulumi-cloudflare/sdk/v5/go/cloudflare"
	"github.com/pulumi/pulumi-digitalocean/sdk/v4/go/digitalocean"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	pulumiConfig "github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

func main() {
	devTestBootstrap2IrohCloudInitYaml, err := os.ReadFile("dev-test-bootstrap2-iroh/cloud-init.yaml")
	if err != nil {
		log.Fatalf("failed to load dev-test-bootstrap2-iroh/cloud-init.yaml: %s", err)
	}

	hcAuthIrohUnytCloudInitBytes, err := os.ReadFile("hc-auth-iroh-unyt/cloud-init.yaml.tmpl")
	if err != nil {
		log.Fatalf("failed to load hc-auth-iroh-unyt/cloud-init.yaml.tmpl: %s", err)
	}
	hcAuthIrohUnytCloudInitTmpl, err := template.New("hc-auth-iroh-unyt-cloud-init").Parse(string(hcAuthIrohUnytCloudInitBytes))
	if err != nil {
		log.Fatalf("failed to parse hc-auth-iroh-unyt/cloud-init.yaml.tmpl: %s", err)
	}

	pulumi.Run(func(ctx *pulumi.Context) error {
		if err := configureDevTestBootstrapSrv(ctx); err != nil {
			return err
		}

		if err := configureDevTestBootstrap2Iroh(ctx, string(devTestBootstrap2IrohCloudInitYaml)); err != nil {
			return err
		}

		if err := configureHcAuthIrohUnyt(ctx, hcAuthIrohUnytCloudInitTmpl); err != nil {
			return err
		}

		return nil
	})
}

func configureDevTestBootstrapSrv(ctx *pulumi.Context) error {
	devTestCloudInitYaml, err := os.ReadFile("dev-test/cloud-init.yaml")
	if err != nil {
		return err
	}

	getSshKeysResult, err := digitalocean.GetSshKeys(ctx, &digitalocean.GetSshKeysArgs{}, nil)
	if err != nil {
		return err
	}

	var sshFingerprints []string
	for _, key := range getSshKeysResult.SshKeys {
		sshFingerprints = append(sshFingerprints, key.Fingerprint)
	}

	_, err = digitalocean.NewDroplet(ctx, "kitsune2-bootstrap-srv", &digitalocean.DropletArgs{
		Image:    pulumi.String("ubuntu-24-04-x64"),
		Name:     pulumi.String("kitsune2-bootstrap-srv"),
		Region:   pulumi.String(digitalocean.RegionFRA1),
		Size:     pulumi.String(digitalocean.DropletSlugDropletS2VCPU2GB),
		Ipv6:     pulumi.Bool(true),
		Tags:     pulumi.StringArray{pulumi.String("network-services")},
		SshKeys:  pulumi.ToStringArray(sshFingerprints),
		UserData: pulumi.String(string(devTestCloudInitYaml)),
	}, pulumi.IgnoreChanges([]string{"sshKeys", "userData"}))
	if err != nil {
		return err
	}

	return nil
}

func configureDevTestBootstrap2Iroh(ctx *pulumi.Context, devTestBootstrap2IrohCloudInitYaml string) error {
	cfg := pulumiConfig.New(ctx, "dns")
	zoneId := cfg.Require("cloudflare-zone-id")

	getSshKeysResult, err := digitalocean.GetSshKeys(ctx, &digitalocean.GetSshKeysArgs{}, nil)
	if err != nil {
		return err
	}

	var sshFingerprints []string
	for _, key := range getSshKeysResult.SshKeys {
		sshFingerprints = append(sshFingerprints, key.Fingerprint)
	}

	droplet, err := digitalocean.NewDroplet(ctx, "dev-test-bootstrap2-iroh", &digitalocean.DropletArgs{
		Image:    pulumi.String("ubuntu-24-04-x64"),
		Name:     pulumi.String("dev-test-bootstrap2-iroh"),
		Region:   pulumi.String(digitalocean.RegionFRA1),
		Size:     pulumi.String(digitalocean.DropletSlugDropletS2VCPU2GB),
		Ipv6:     pulumi.Bool(true),
		Tags:     pulumi.StringArray{pulumi.String("network-services")},
		SshKeys:  pulumi.ToStringArray(sshFingerprints),
		UserData: pulumi.String(devTestBootstrap2IrohCloudInitYaml),
	}, pulumi.IgnoreChanges([]string{"sshKeys"}))
	if err != nil {
		return err
	}

	_, err = cloudflare.NewRecord(ctx, "dev-test-bootstrap2-iroh-A", &cloudflare.RecordArgs{
		ZoneId:  pulumi.String(zoneId),
		Name:    pulumi.String("dev-test-bootstrap2-iroh"),
		Type:    pulumi.String("A"),
		Content: droplet.Ipv4Address,
		Ttl:     pulumi.Int(300),
		Proxied: pulumi.Bool(false),
	})
	if err != nil {
		return err
	}

	_, err = cloudflare.NewRecord(ctx, "dev-test-bootstrap2-iroh-AAAA", &cloudflare.RecordArgs{
		ZoneId:  pulumi.String(zoneId),
		Name:    pulumi.String("dev-test-bootstrap2-iroh"),
		Type:    pulumi.String("AAAA"),
		Content: droplet.Ipv6Address,
		Ttl:     pulumi.Int(300),
		Proxied: pulumi.Bool(false),
	})
	if err != nil {
		return err
	}

	return nil
}

func configureHcAuthIrohUnyt(ctx *pulumi.Context, cloudInitTmpl *template.Template) error {
	getSshKeysResult, err := digitalocean.GetSshKeys(ctx, &digitalocean.GetSshKeysArgs{}, nil)
	if err != nil {
		return err
	}

	var sshFingerprints []string
	for _, key := range getSshKeysResult.SshKeys {
		sshFingerprints = append(sshFingerprints, key.Fingerprint)
	}

	cfg := pulumiConfig.New(ctx, "hc-auth-iroh-unyt")
	githubClientId := cfg.RequireSecret("github-client-id")
	githubClientSecret := cfg.RequireSecret("github-client-secret")
	sessionSecret := cfg.RequireSecret("session-secret")
	apiTokens := cfg.RequireSecret("api-tokens")

	userData := pulumi.All(githubClientId, githubClientSecret, sessionSecret, apiTokens).ApplyT(
		func(args []interface{}) (string, error) {
			data := map[string]string{
				"GithubClientId":     args[0].(string),
				"GithubClientSecret": args[1].(string),
				"SessionSecret":      args[2].(string),
				"ApiTokens":          args[3].(string),
			}
			var buf bytes.Buffer
			if err := cloudInitTmpl.Execute(&buf, data); err != nil {
				return "", err
			}
			return buf.String(), nil
		}).(pulumi.StringOutput)

	_, err = digitalocean.NewDroplet(ctx, "hc-auth-iroh-unyt", &digitalocean.DropletArgs{
		Image:    pulumi.String("ubuntu-24-04-x64"),
		Name:     pulumi.String("hc-auth-iroh-unyt"),
		Region:   pulumi.String(digitalocean.RegionFRA1),
		Size:     pulumi.String(digitalocean.DropletSlugDropletS4VCPU8GB),
		Ipv6:     pulumi.Bool(true),
		Tags:     pulumi.StringArray{pulumi.String("network-services")},
		SshKeys:  pulumi.ToStringArray(sshFingerprints),
		UserData: userData,
	}, pulumi.IgnoreChanges([]string{"sshKeys", "userData"}))
	return err
}
