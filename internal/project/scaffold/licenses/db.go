package licenses

import (
	"embed"
	"fmt"
)

//go:embed sync/*.txt
var db embed.FS

// License defines a software license that can be included in the repository. It contains a FullName that is used for
// display, an SPDX Identifier, and a TemplateName reference. In addition to the metadata, we also track
type License struct {
	FullName     string
	Identifier   string
	Free         bool
	OSIApproved  bool
	TemplateName string
}

// Text attempts to read the license text from the embedded db.
func (l License) Text() (string, bool) {
	body, err := db.ReadFile(fmt.Sprintf("sync/%s.txt", l.TemplateName))
	if err != nil {
		return "", false
	}

	return string(body), true
}

var (
	//BuiltIn defines the list of licenses that are built into the tool.
	BuiltIn = []License{
		{"GNU Affero General Public License v3.0 only", "AGPL-3.0-only", true, true, "agpl3"},
		{"Apache License 2.0", "Apache-2.0", true, true, "apache"},
		{"MIT License", "MIT", true, true, "mit"},
		{"Mozilla Public License 2.0", "MPL-2.0", true, true, "mpl"},
	}

	// BySPDX indexes the BuiltIn Licenses by their SPDX Identifier.
	BySPDX = map[string]License{}

	// ByTemplateName indexes the BuiltIn Licenses by their template names.
	ByTemplateName = map[string]License{}
)

func init() {
	for _, builtin := range BuiltIn {
		BySPDX[builtin.Identifier] = builtin
		ByTemplateName[builtin.TemplateName] = builtin
	}
}
