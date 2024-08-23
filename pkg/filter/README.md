# Filter

The filter package is designed to allow you to define filters, primarily by yaml to be able to filter resources. This is
used in the `nuke` package to filter resources based on a set of criteria.

Filter's can be optionally added to a `group` filters within a `group` are combined with an `AND` operation. Filters in
different `group` are combined with an `OR` operation.

There is also the concept of a `global` filter that is applied to all resources.

## Types

There are multiple filter types that can be used to filter the resources. These types are used to match against the
property.

- empty
- exact
- glob
- regex
- contains
- dateOlderThan
- suffix
- prefix

## Global

You can define a global filter that will be applied to all resources. This is useful for defining a set of filters that
should be applied to all resources.

It has a special key called `__global__`.

This only works when you are defining it as a resource type as part of the `Filters` `map[string][]Filter` type.
