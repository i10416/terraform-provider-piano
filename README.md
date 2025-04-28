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
go tool github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs generate --provider-dir . -provider-name piano
```

### Misc

```sh
go tool github.com/hashicorp/copywrite headers -d . --config ./.copywrite.hcl
```
