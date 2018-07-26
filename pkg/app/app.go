package app

import (
	"confmaster/pkg/kv"
	"confmaster/pkg/template"
)

// app is an application that contains one configuration backed by single kvstore
type App interface {

	Configs() kv.SnapshotStore

	Templates() template.SnapshotStore

	Labels() Labels

	Attributes() Attributes

	Engine() template.Engine
}
