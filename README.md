# libnuke

[![License](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)
[![GoDoc](https://godoc.org/github.com/ekristen/libnuke?status.svg)](https://godoc.org/github.com/ekristen/libnuke)
[![Go Report Card](https://goreportcard.com/badge/github.com/ekristen/libnuke)](https://goreportcard.com/report/github.com/ekristen/libnuke)
[![Maintainability](https://api.codeclimate.com/v1/badges/dc4078a236e89486b4ca/maintainability)](https://codeclimate.com/github/ekristen/libnuke/maintainability)
[![codecov](https://codecov.io/gh/ekristen/libnuke/graph/badge.svg?token=UJJOUJ98G4)](https://codecov.io/gh/ekristen/libnuke)
[![tests](https://github.com/ekristen/libnuke/actions/workflows/tests.yml/badge.svg)](https://github.com/ekristen/libnuke/actions/workflows/tests.yml)

**Status: [Initial Development](https://semver.org/spec/v2.0.0-rc.1.html#spec-item-5)** - Everything works, but is still being abstracted and tailored to aws-nuke and azure-nuke,
as such func signatures and other things may change in breaking ways until things stabilize.

## Overview

This is an attempt to consolidate the commonalities between [aws-nuke](https://github.com/ekristen/aws-nuke) and [azure-nuke](https://github.com/ekristen/azure-nuke) into a single library
that can be used between them and for future tooling, for example [gcp-nuke](https://github.com/ekristen/gcp-nuke). Additionally, the goal is to make it
easier to add new features with better test coverage.

The goal of this library is to have a well tested and stable library to build additional nuke tools on top of, while
reducing the technical debt overhead of managing each tool individually. By abstracting away and testing the core parts
of the code, each implementing tool can focus on adding resources to remove and lower the barrier of entry for new
contributors.

## Attribution, License, and Copyright

First of tall this library would not have been possible without the hard work of the team over at [rebuy-de](https://github.com/rebuy-de)
and their original work on [rebuy-de/aws-nuke](https://github.com/rebuy-de/aws-nuke).

This library is licensed under the MIT license. See the [LICENSE](LICENSE) file for more information. The bulk of this
library was originally sourced from [rebuy-de/aws-nuke](https://github.com/rebuy-de/aws-nuke). See the [Sources](#sources)
for more.

## History of the Library

This all started when I created a managed fork of [aws-nuke](https://github.com/ekristen/aws-nuke) from the [original aws nuke](https://github.com/rebuy-de/aws-nuke).
The fork become necessary after attempting to make contributions and respond to issues to learn that the current 
maintainers only have time to work on the project about once a month and while receptive to bringing in other people
to help maintain, made it clear it would take time. Considering the feedback cycle was already weeks on initial
communications, I had to make the hard decision to fork and maintain myself.

After the fork, I created [azure-nuke](https://github.com/ekristen/azure-nuke) to fulfill a missing need there and 
quickly realized that it would be great to pull all the common code into a common library that could be shared by the
two tools with the realization I would be also be making [gcp-nuke](https://github.com/ekristen/gcp-nuke) in the near
future.

### A Few Note About the Original Code

The code that was originally written for [aws-nuke](https://github.com/rebuy-de/aws-nuke) for iterating over and clearing out resources was well 
written and I wanted to be able to use it for other cloud providers. Originally I simply copied it for [azure-nuke,](https://github.com/ekristen/azure-nuke) 
but I didn't want to have to keep on maintaining multiple copies.

There are a few shortcomings with the original code base, for example, there's no way to do dependency management. For 
example there are some resources that must be cleared before other resources can be cleared, or it will end in error. Now
the retry mechanism is **usually** sufficient for this, but not always.

The queue code in my opinion was very novel in its approach and I wanted to keep that, but I wanted to make sure it was
agnostic to the system using it. As such, the queue package can be used for just about anything in which you want to queue
and retry items. However, it is still geared towards the removal of said it, it's primary interface has to have the
`Remove` method still available.

## License

MIT

## Sources

Most of this code originated from the original [aws-nuke](https://github.com/rebuy-de/aws-nuke) project.

- [aws-nuke](https://github.com/ekristen/aws-nuke) (managed fork)
- [aws-nuke original](https://github.com/rebuy-de/aws-nuke)
- [azure-nuke](https://github.com/ekristen/azure-nuke)

## Versioning

This library will follow the semver model. However, it is still in alpha/beta and as such the API is subject to change
until it is stable and will remain on the `0.y.z` model until then.

## Packages

I strongly dislike the use of the `internal` directory in any open source golang project. Therefore, everything is in
the `pkg` directory and exported wherever possible to allow others to use it. This project follows the semver model, so
breaking changes will be made in a way that is compatible with semver.

### config

This provides the configuration for libnuke. It contains the configuration for all the accounts, regions,
and resource types. It also contains the presets that can be used to apply a set of filters to a nuke process. The
configuration is loaded from a YAML file and is meant to be used by the implementing tool. Use of the configuration
is not required but is recommended. The configuration can be implemented a specific way for each tool providing it
has the necessary methods available.

### errors

This provides common errors that can be used throughout the library for handling of resource errors

### filter

This provides a way to filter resources based on a set of criteria.

### log

This is a simple wrapper around `fmt.Println` that formats resource cleanup messages nicely.

### nuke

This provides the framework for scanning for resources and then iterating over said resources to determine
if they should be removed or not and in what order.

### queue

This provides a simple list mechanism with some helper functions to determine current counts based on
resource type or state.

### resource

This provides a way to interact with resources. This provides multiple interfaces to test against
as resources can optionally implement these interfaces.

### settings

This provides a way to handle settings for the library. The primary use case is arbitrary settings that resources might
need to be configurable that changes the behavior of how said resource is to be removed. For example, EC2Instances
have Deletion Protection, this allows the resource to define it needs a setting called `DisableDeletionProtection` and then 
allows that to be defined in the `config` package and then passed to the resource when it is being removed.

### types

This is a collection of common types that are used throughout the library.

### utils

This is a collection of common utilities that are used throughout the library.