# Filter

The filter package is designed to allow you to define filters, primarily by yaml to be able to filter resources. This is
used in the `nuke` package to filter resources based on a set of criteria.

Filter's can be optionally added to a `group` filters within a `group` are combined with an `AND` operation. Filters in
different `group` are combined with an `OR` operation.

There is also the concept of a `global` filter that is applied to all resources.

## Global

You can define a global filter that will be applied to all resources. This is useful for defining a set of filters that
should be applied to all resources.

It has a special key called `__global__`.

This only works when you are defining it as a resource type as part of the `Filters` `map[string][]Filter` type.

## Types

There are multiple filter types that can be used to filter the resources. These types are used to match against the
property.

- empty
- exact
- glob
- regex
- contains
- dateOlderThan
- dateOlderThanNow
- suffix
- prefix
- In
- NotIn

### empty / exact

These are identical, if you leave your type empty, it will choose exact. Exact will only match if values are identical.

### glob

A glob allows for matching values using asterisk for a wild card, you may have more than one asterisk.

### regex

A regex allows for matching values with any valid regular expression.

### contains

A contains type allows for matching a value if it has the value contained within the property value.

### dateOlderThan

This allows you to filter a property's value based on whether it is older 

### dateOlderThanNow

This allows you to filter a properties value based on whether it is older than the current time in UTC with an addition
or subtraction of a duration value.

For example if the property is `CreatedDate` and the value is `2024-04-14T12:00:00Z` and the current time is
`2024-04-14T18:00:00Z` then you can set a negative duration like `-12h`. In this case it would not match, as the
`CreatedDate` would be **after** the adjusted time. 

If you adjusted it `-4h` then it **would** match as the `CreatedDate` would be **before** the adjusted time.

### suffix

This allows you to match a value if the value being filtered on is the suffix of the property value.

### prefix

This allows you to match a value if the value being filtered on is the prefix of the property value.

### In

This allows you to match a value if it is in a list of values.

### NotIn

This allows you to match a value if it is not in a list of values.