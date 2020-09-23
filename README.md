# terraform-registry
This is a light weight Terraform Registry, more like a proxy.
It currently only supports the `v1.provider` endpoint and Terraform provider releases hosted on Github.

# how it works
The registry dynamically generates the correct response based on assets found in
Github provider releases which conform to the Terraform asset conventions.
There is one additional file required which should be called `signkey.asc`
This file must contain the [ASCII Armored PGP public key](https://www.terraform.io/docs/registry/providers/publishing.html) which was
used to sign the `..._SHA256SUMS.sig` signature file.

# use cases
- host your own private Terraform provider registry
- easily release custom builds of providers e.g. releases from your own forks

# deployment
Build a docker image and deploy it to your favorite hosting location

# endpoints
| Endpoint | Description |
|-----------|-------------|
| `/.well-known/terraform.json` | The service discovery endpoint used by terraform |
| `/v1/providers/:namespace/:type/*` | The `versions` and `download` action endpoints |

# example usage

```terraform
terraform {
    required_providers {
      cloudfoundry = {
        source  = "terraform-registry.us-east.philips-healthsuite.com/philips-forks/cloudfoundry"
        version = "0.12.2-202008131826"
      }
    }
  }
}
```

The above assumes a copy of the terraform-registry running at:

https://terraform-registry.us-east.philips-healthsuite.com

It references `philips-forks/cloudfoundry` version `0.12.2-202008131826` which maps to the following Github repository and release:

https://github.com/philips-forks/terraform-provider-cloudfoundry/releases/tag/v0.12.2-202008131826

Notice the `signkey.asc` which is included in this release. You can use [Goreleaser](https://goreleaser.com/quick-start/) with this [.goreleaser.yml](https://github.com/hashicorp/terraform-provider-scaffolding/blob/master/.goreleaser.yml) template to create arbitrary releases of providers. The provider pointer also does not include the `terraform-provider-` prefix.

# private repositories

## authenticating via Personal Access Token

1. Create token with `repo` scope [here](https://github.com/settings/tokens/new)

  If you are using GitHub SSO for your organization, press `Enable SSO` button on your token and authorize it for this organization.

2. Set token in `GITHUB_TOKEN` environment variable

# current limitations and TODOs
- Only supports providers

# contact / getting help
andy.lo-a-foe@philips.com

# license
License is MIT
