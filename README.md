# Piano.io Terraform Provider

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.24

## Building The Provider

1. Clone the repository
1. Enter the repository directory
1. Build the provider using the Go `install` command:

```shell
go install
```

## Using the provider

```tf
provider "piano" {
  endpoint  = "https://sandbox.piano.io/api/v3"
  api_token = "*********************"
}
```

For more details, visit https://registry.terraform.io/providers/i10416/piano/latest/docs

## Developing the provider

### Debugging

#### Setup .terraform.tfrc

Create `.terraform.tfrc` file so that terraform can lookup local provider implementation.

```
provider_installation {

  dev_overrides {
      "hashicorp.com/i10416/piano" = "$GOBIN"
  }

  direct {}
}

```

#### Install go binary

In workspace root:

```sh
go install .
```

This will install go binary under GOBIN.

#### Running local provider implementation
Now, terraform picks up the local provider implementation when you run `terraform` command.

```tf
terraform {
  required_providers {
    piano = {
      source = "hashicorp.com/i10416/piano"
    }
  }
}
```

```sh
terraform <cmd>
```

For more details, check https://developer.hashicorp.com/terraform/tutorials/providers-plugin-framework/providers-plugin-framework-provider#prepare-terraform-for-local-provider-install

#### Generating piano SDK for Go

This repository contains internal piano publisher SDK generated from piano.io OpenAPI specification with some modification.

Run the following command to keep generated SDK in sync with OpenAPI spec.

```sh
go generate ./.
```

### Docs

In project root:

```sh
go tool tfplugindocs generate --provider-dir . -provider-name piano
```

### Misc

```sh
go tool copywrite headers -d . --config ./.copywrite.hcl
```
