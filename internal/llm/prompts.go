package llm

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/ol1n/auction-sim/internal/auction"
)

// Registry drží načtené prompt šablony z adresáře (role_*.md, decision_*.md).
type Registry struct {
	roles     map[string]string
	decisions map[auction.AuctionType]*template.Template
}

// LoadRegistry načte všechny prompt soubory z dir.
func LoadRegistry(dir string) (*Registry, error) {
	r := &Registry{
		roles:     map[string]string{},
		decisions: map[auction.AuctionType]*template.Template{},
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("čtení prompts dir: %w", err)
	}
	for _, e := range entries {
		name := e.Name()
		if !strings.HasSuffix(name, ".md") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return nil, err
		}
		switch {
		case strings.HasPrefix(name, "role_"):
			role := strings.ToUpper(strings.TrimSuffix(strings.TrimPrefix(name, "role_"), ".md"))
			r.roles[role] = string(data)
		case strings.HasPrefix(name, "decision_"):
			typ := auction.AuctionType(strings.ToUpper(strings.TrimSuffix(strings.TrimPrefix(name, "decision_"), ".md")))
			tmpl, err := template.New(name).Parse(string(data))
			if err != nil {
				return nil, fmt.Errorf("parse %s: %w", name, err)
			}
			r.decisions[typ] = tmpl
		}
	}
	return r, nil
}

// RoleSystem vrátí system prompt pro roli (prázdné, pokud chybí).
func (r *Registry) RoleSystem(role string) string {
	return r.roles[strings.ToUpper(role)]
}

// RenderDecision vyrenderuje decision prompt pro daný typ aukce s předanými daty.
func (r *Registry) RenderDecision(t auction.AuctionType, data any) (string, error) {
	tmpl, ok := r.decisions[t]
	if !ok {
		return "", fmt.Errorf("chybí decision šablona pro %s", t)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
