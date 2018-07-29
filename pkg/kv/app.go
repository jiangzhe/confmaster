package kv

import (
	"confmaster/pkg/template"
)

// app is an application that contains one configuration backed by single kvstore
type App interface {

	Configs() SnapshotStore

	Templates() template.SnapshotStore

	Labels() Labels

	Attributes() Attributes

	Engine() template.Engine
}
