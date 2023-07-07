package transformer

import (
	"encoding/json"
	"strings"
	"unicode"

	"github.com/ergomake/ergomake/internal/database"
)

type EnvironmentService struct {
	ID            string            `json:"-"`
	Url           string            `json:"url"`
	Image         string            `json:"image"`
	Build         string            `json:"build"`
	Index         int               `json:"index"`
	PublicPort    string            `json:"-"`
	InternalPorts []string          `json:"-"`
	Env           map[string]string `json:"-"`
}

type Environment struct {
	Services   map[string]EnvironmentService `json:"services"`
	RawContent string                        `json:"-"`
}

func (c *Environment) ToMap() map[string]interface{} {
	b, _ := json.Marshal(c)
	var m map[string]interface{}
	_ = json.Unmarshal(b, &m)

	m["apps"] = m["services"]

	return m
}

func EnvironmentFromDB(env *database.Environment) *Environment {
	services := make(map[string]EnvironmentService)
	for _, svc := range env.Services {
		services[svc.Name] = EnvironmentService{
			ID:            svc.ID,
			Url:           svc.Url,
			Image:         svc.Image,
			Build:         svc.Build,
			Index:         svc.Index,
			PublicPort:    svc.PublicPort,
			InternalPorts: svc.InternalPorts,
		}
	}

	return &Environment{
		Services: services,
	}
}

func NewEnvironment(services map[string]EnvironmentService, rawCompose string) *Environment {
	c := Environment{Services: services, RawContent: rawCompose}
	c.computeServicesIndexes()
	return &c
}

func (c *Environment) computeServicesIndexes() {
	lines := strings.Split(c.RawContent, "\n")

	insideServices := false
	index := 0
	servicesIdentation := -1
	for _, line := range lines {
		if strings.HasPrefix(line, "services:") {
			insideServices = true
			continue
		}

		if !insideServices {
			continue
		}

		var first rune
		for _, c := range line {
			first = c
			break
		}
		if len(line) != 0 && !unicode.IsSpace(first) {
			// when line is not empty but does not starts with whitespace
			// it means we are out of the services block
			break
		}

		trimmedLine := strings.TrimSpace(line)
		if strings.HasPrefix(trimmedLine, "#") || len(trimmedLine) == 0 {
			continue
		}

		currentIdentation := len(line) - len(trimmedLine)
		if servicesIdentation == -1 {
			// if this is not a comment and identation is not set, we must have found the first service definition
			// so we store the identation and from now on, we only care about lines that have the same identation
			servicesIdentation = currentIdentation
		}

		if currentIdentation != servicesIdentation {
			continue
		}

		parts := strings.SplitN(trimmedLine, ":", 2)
		name := strings.TrimSpace(parts[0])

		svc, ok := c.Services[name]
		if ok {
			svc.Index = index
			index += 1
			c.Services[name] = svc
		}
	}
}

func (c *Environment) FirstService() EnvironmentService {
	service := EnvironmentService{}
	for _, s := range c.Services {
		if s.Index == 0 {
			service = s
		}
	}

	return service
}
