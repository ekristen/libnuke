package types

// ResolveResourceTypes resolves the resource types based on the provided includes, excludes, and alternatives.
// The alternatives are a list of resource types that are to be used instead of the default resource. The primary use
// case for this is AWS Cloud Control API resources. If a resource has been registered with the Cloud Control API.
// Includes, Excludes, and Alternatives are []Collection because they are a combination of runtime, global and account
// level configuration.
func ResolveResourceTypes(
	base Collection,
	includes, excludes, alternatives []Collection,
	alternativeMappings map[string]string) Collection {

	for _, cl := range alternatives {
		oldStyle := Collection{}
		for _, c := range cl {
			os, found := alternativeMappings[c]
			if found {
				oldStyle = append(oldStyle, os)
			}
		}

		base = base.Union(cl)
		base = base.Remove(oldStyle)
	}

	for _, i := range includes {
		if len(i) > 0 {
			base = base.Intersect(i)
		}
	}

	for _, e := range excludes {
		base = base.Remove(e)
	}

	return base
}
