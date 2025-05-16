# terra-profile

Automatically set the respective `AWS_PROFILE` before calling `terraform` or `terragrunt` based on an `account.hcl` file
exporting a local profile name.

Never accidentally plan or apply against the wrong account because you forgot to change the `AWS_PROFILE`.

## Example usage

This CLI tool will search backwards within a directory and find the first `account.hcl` file and read the `account_name`
from locals, for example:

```bash
locals {
  account_name   = "my-prod-account"
  aws_account_id = "0123456789012"
}
```

This would do the equivalent of:

```bash
AWS_PROFILE=my-prod-account terragrunt run-all apply
```
