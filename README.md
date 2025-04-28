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

In workspace root:

```sh
go install .
```

In example/provider-interactive-debugging:

```sh
terraform <cmd>
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
