# Tags Module

Configurations in this module creates tags with default values that's
suitable for flux tests.

## Usage

```hcl
module "tags" {
    source = "git::https://github.com/fluxcd/test-infra.git//tf-modules/utils/tags"

    tags = {
        Platform = "Foo"
    }
}
```

**NOTE:** When using this module with the terraform AWS provider for setting the
resource tags https://github.com/hashicorp/terraform-provider-aws/issues/19583
issue has been observed which happens when using any tag with dynamic value
that's determined by terraform at runtime. A workaround is to overwrite the
default value of `createdat` tag by setting it as an input. When setting the
tags through environment variables, following can be used to create an override
value:

```sh
export TF_VAR_tags='{"dev"="true", "createdat"='"\"$(date -u +x%Y-%m-%d_%Hh%Mm%Ss)\""'}'
```

The same issue also happens if the name of the resource is dynamically
constructed in presence of any tags, static or dynamic. Use static resource
name, for example by setting a dynamically generated value in environment variable using
`$RANDOM`, to prevent such issues.
For example, setting variable as
```sh
export TF_VAR_rand=${RANDOM}
```
and using it in config
```hcl
locals {
  name = "flux-test-${var.rand}"
}
```
