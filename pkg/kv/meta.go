package kv

import (
	"errors"
	"confmaster/pkg/template"
)

var (
	ErrMultipleApps = errors.New("multiple app found")
	ErrAppNotExists = errors.New("app not exists")
)

// app is an application that contains one configuration backed by single kvstore
type App interface {

	Configs() SnapshotStore

	Templates() template.SnapshotStore

	Labels() Labels

	Attributes() Attributes

	Engine() template.Engine
}

// AppLabels are associated to an app, and can be part of search patterns
// the labels are immutable during whole app lifecycle,
// it's more like a schema/structure information,
// for mutable attributes, use AppAttributes
type Labels map[string]string

// AppAttributes are also associated to an app, but can be changed
type Attributes map[string]string

// AppLocator can locate single app, or search multiple apps with patterns
type Locator interface {
	// locate single App, if the Namespace and labels match more than one app,
	// it will return ErrMultipleApps
	Locate(Namespace, Labels) (App, error)
	Search(Namespace, Labels, Attributes) []App
}

// namespace is concept of a separated area of apps, maybe with different
// resource quota, user accounts, access control, etc.
// this is similar with Kubernetes "namespace"
type Namespace string

const (
	NamespaceAll Namespace = ""
)

