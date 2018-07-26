package dep

import "confmaster/pkg/app"

// a dependency is a one-to-one link that specify one app will
// reuse part of another app's config
type Dependency interface {
	// upstream labels can locate single app as upstream
	Upstream() app.Labels

	// downstream labels can locate single app as downstream
	Downstream() app.Labels

	// get the mapping between two config paths
	// a meaningful dependency should have at least one mapping
	Mappings() []PathMapping
}

// a path mapping specify how part of upstream config map
// to part of downstream config
type PathMapping struct {
	Source []string
	Target []string
}