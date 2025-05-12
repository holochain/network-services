package main

import (
	"github.com/pulumi/pulumi-digitalocean/sdk/v4/go/digitalocean"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"log"
	"os"
)

func main() {
	devTestCloudInitYaml, err := os.ReadFile("dev-test/cloud-init.yaml")
	if err != nil {
		log.Fatalf("failed to load cloud-init.yaml: %s", err)
	}

	devTestAuthCloudInitYaml, err := os.ReadFile("dev-test-auth/cloud-init.yaml")
	if err != nil {
		log.Fatalf("failed to load cloud-init.yaml: %s", err)
	}

	pulumi.Run(func(ctx *pulumi.Context) error {
		if err := configureDevTestBootstrapSrv(ctx, string(devTestCloudInitYaml)); err != nil {
			return err
		}

		if err := configureDevTestAuthBootstrapSrv(ctx, string(devTestAuthCloudInitYaml)); err != nil {
			return err
		}

		return nil
	})
}

func configureDevTestBootstrapSrv(ctx *pulumi.Context, devTestCloudInitYaml string) error {
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
		UserData: pulumi.String(devTestCloudInitYaml),
	}, pulumi.IgnoreChanges([]string{"SshKeys"}))
	if err != nil {
		return err
	}

	return nil
}

func configureDevTestAuthBootstrapSrv(ctx *pulumi.Context, devTestAuthCloudInitYaml string) error {
	getSshKeysResult, err := digitalocean.GetSshKeys(ctx, &digitalocean.GetSshKeysArgs{}, nil)
	if err != nil {
		return err
	}

	var sshFingerprints []string
	for _, key := range getSshKeysResult.SshKeys {
		sshFingerprints = append(sshFingerprints, key.Fingerprint)
	}

	_, err = digitalocean.NewDroplet(ctx, "kitsune2-bootstrap-srv-auth", &digitalocean.DropletArgs{
		Image:    pulumi.String("ubuntu-24-04-x64"),
		Name:     pulumi.String("kitsune2-bootstrap-srv-auth"),
		Region:   pulumi.String(digitalocean.RegionFRA1),
		Size:     pulumi.String(digitalocean.DropletSlugDropletS2VCPU2GB),
		Ipv6:     pulumi.Bool(true),
		Tags:     pulumi.StringArray{pulumi.String("network-services")},
		SshKeys:  pulumi.ToStringArray(sshFingerprints),
		UserData: pulumi.String(devTestAuthCloudInitYaml),
	}, pulumi.IgnoreChanges([]string{"SshKeys"}))
	if err != nil {
		return err
	}

	return nil
}
