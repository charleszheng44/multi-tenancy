/*
Copyright 2021 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package plugin

import (
	"sync"

	pkgerr "github.com/pkg/errors"
)

var (
	// ErrNoPluginID is returned when no id is specified
	ErrNoPluginID = pkgerr.New("plugin: no id")
)

// Registration contains information for registering a plugin
type Registration struct {
	// ID of the plugin
	ID string

	// InitFn is called when initializing a plugin. The registration and
	// context are passed in.
	InitFn func(*InitContext) (interface{}, error)
	// Disable the plugin from loading
	Disable bool
}

// Init the registered plugin
func (r *Registration) Init(ic *InitContext) *Plugin {
	p, err := r.InitFn(ic)
	return &Plugin{
		Registration: r,
		instance:     p,
		err:          err,
	}
}

// Plugin represents an initialized plugin, used with an init context.
type Plugin struct {
	Registration *Registration // registration, as initialized

	instance interface{}
	err      error // will be set if there was an error initializing the plugin
}

// Instance returns the instance and any initialization error of the plugin
func (p *Plugin) Instance() (interface{}, error) {
	return p.instance, p.err
}

var register = struct {
	sync.RWMutex
	r map[string]*Registration
}{}

// Register allows plugins to register
func Register(r *Registration) {
	register.Lock()
	defer register.Unlock()
	if r.ID == "" {
		panic(ErrNoPluginID)
	}

	if register.r == nil {
		register.r = make(map[string]*Registration)
	}

	register.r[r.ID] = r
}

// List returns the list of registered plugins for initialization.
func List() []*Registration {
	var res []*Registration
	register.RLock()
	defer register.RUnlock()
	for id := range register.r {
		res = append(res, register.r[id])
	}

	return res
}
