/*
Copyright 2019 Banzai Cloud.

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

package remoteclusters

import (
	"errors"
	"sync"
)

type Manager struct {
	clusters map[string]*Cluster
	mu       *sync.RWMutex
}

func NewManager() *Manager {
	return &Manager{
		clusters: make(map[string]*Cluster),
		mu:       &sync.RWMutex{},
	}
}

func (m *Manager) GetAll() map[string]*Cluster {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.clusters
}

func (m *Manager) Add(cluster *Cluster) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.clusters[cluster.GetName()] = cluster

	return nil
}

func (m *Manager) Delete(cluster *Cluster) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.clusters, cluster.GetName())

	return nil
}

func (m *Manager) Get(name string) (*Cluster, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	cluster := m.clusters[name]
	if cluster == nil {
		return nil, errors.New("cluster not found")
	}

	return cluster, nil
}
