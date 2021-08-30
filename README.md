# Terraform Provider - Custom AWS stuff

This is a first implementation of providers to add some missing functionality to the AWS provider.
Don't expect it to be perfect:

- first try with Go development :)
- first try with terraform plugin development :)

## Build provider

Run the following command to build the provider

```shell
$ go build -o terraform-provider-hashicups
```

## Test sample configuration

First, build and install the provider.

```shell
$ make install
```

Then, navigate to the `examples` directory. 

```shell
$ cd examples
```

Run the following command to initialize the workspace and apply the sample configuration.

```shell
$ terraform init && terraform apply
```
