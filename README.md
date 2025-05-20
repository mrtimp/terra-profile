# terra-profile

Automatically set the respective `AWS_PROFILE` before calling `terragrunt` based on an `account.hcl` file
exporting a local profile name.

Never accidentally plan or apply against the wrong account because you forgot to change the `AWS_PROFILE`.

By default, the --provider-cache option will be enabled, this is helpful by cutting down on storage when working on
larger Terraform projects. You can disable this with the `--disable-provider-cache` option.

## Example usage

This CLI tool will search backwards within a directory and find the first `account.hcl` file and read the `account_name`
from locals, for example:

Having the following account.hcl file within your Terragrunt project tree:

```bash
locals {
  account_name   = "my-prod-account"
  aws_account_id = "0123456789012"
}
```

and executing:

```bash
terra-profile [--debug] [--non-interactive] [--disable-provider-cache] terragrunt run-all apply
```

would do the equivalent of:

```bash
AWS_PROFILE=my-prod-account terragrunt run-all [--provider-cache] [--non-interactive] apply
```
