package templates

import (
	"nems/internal/models"
	"sort"
	"sync"
)

type TemplateMetadata struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"` // "modbus", "rest", "cloud", "cloud_rest", "demo"
}

type Template struct {
	Metadata  TemplateMetadata
	NewPoller func(device models.Device) models.DevicePoller
}

var (
	mu        sync.RWMutex
	templates = make(map[string]Template)
)

func Register(t Template) {
	mu.Lock()
	defer mu.Unlock()
	templates[t.Metadata.ID] = t
}

func GetTemplates() []TemplateMetadata {
	mu.RLock()
	defer mu.RUnlock()

	var list []TemplateMetadata
	for _, t := range templates {
		list = append(list, t.Metadata)
	}

	sort.Slice(list, func(i, j int) bool {
		return list[i].Name < list[j].Name
	})

	return list
}

func CreatePoller(id string, device models.Device) models.DevicePoller {
	mu.RLock()
	defer mu.RUnlock()

	if t, ok := templates[id]; ok {
		return t.NewPoller(device)
	}
	return nil
}
