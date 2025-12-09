package templates

var templates = map[string]string{
	"naming": `**Naming Convention**

Consider using a more descriptive name that follows the project's naming conventions.`,

	"security": `**Security Concern**

This code may have security implications. Please review for potential vulnerabilities.`,

	"perf": `**Performance**

This implementation may have performance implications at scale. Consider optimizing.`,

	"style": `**Style Guide**

This code doesn't follow the project's style guide. Please update to match conventions.`,
}

func Get(name string) (string, bool) {
	template, ok := templates[name]
	return template, ok
}

func List() []string {
	names := make([]string, 0, len(templates))
	for name := range templates {
		names = append(names, name)
	}
	return names
}
