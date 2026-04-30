package main

import (
	"fmt"
	"sort"
)

// impModel is used for storing import information
type impModel struct {
	path           string
	localReference string
}

// string is used to get a string representation of an import
func (m impModel) string() string {
	if m.localReference == "" {
		return m.path
	}

	return m.localReference + " " + m.path
}

const (
	GroupStandard int = iota // 0
	GroupThird
	GroupSecond
	GroupLocal
	GroupCount
)

type impManager struct {
	groups []*impGroup
}

type impGroup struct {
	models []*impModel
}

func (g *impGroup) append(model *impModel) {
	g.models = append(g.models, model)
}

func newImpManager() *impManager {
	groups := make([]*impGroup, GroupCount)
	for idx := range groups {
		groups[idx] = &impGroup{
			models: []*impModel{},
		}
	}
	return &impManager{groups: groups}
}

func (m *impManager) Standard() *impGroup {
	return m.groups[GroupStandard]
}

func (m *impManager) Local() *impGroup {
	return m.groups[GroupLocal]
}

func (m *impManager) ThirdPart() *impGroup {
	return m.groups[GroupThird]
}

func (m *impManager) SecondPart() *impGroup {
	return m.groups[GroupSecond]
}

func (m *impManager) sortImports() {
	for _, g := range m.groups {
		g.sortImports()
	}
}

// sortImports sorts multiple imports by import name & prefix
func (g *impGroup) sortImports() {
	imports := g.models
	sort.Slice(imports, func(i, j int) bool {
		if imports[i].path != imports[j].path {
			return imports[i].path < imports[j].path
		}
		return imports[i].localReference < imports[j].localReference
	})
}

// convertImportsToGo generates output for correct categorised import statements
func (m *impManager) convertImportsToGo() []byte {
	output := "import ("

	for _, group := range m.groups {
		if group.countImports() == 0 {
			continue
		}
		output += "\n"
		for _, imp := range group.models {
			output += fmt.Sprintf("\t%v\n", imp.string())
		}
	}

	output += ")"

	return []byte(output)
}

func (g *impGroup) countImports() int {
	return len(g.models)
}

// countImports count the total number of imports of a [][]impModel
func (m *impManager) countImports() int {
	count := 0
	for _, group := range m.groups {
		count += group.countImports()
	}
	return count
}
