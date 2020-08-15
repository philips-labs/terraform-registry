# terraform-registry
This is light weight Terraform Registry, more like a proxy.
It currently only supports a provider endpoint and Github Terraform provider releases.

# how it works
The registry dynamically generates the correct response from
Github provider releases which conform to the Terraform asset releases
There is one additional file required which should be called `signkey.asc`
This file should contain the ASCII Armored PGP public key which was
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

# limitations and TODOs
- Uses an anonymous Github client which has very low quota
- Only supports providers
- TODO: support Github PAT tokens

# contact / getting help
andy.lo-a-foe@philips.com

# license
License is MIT
