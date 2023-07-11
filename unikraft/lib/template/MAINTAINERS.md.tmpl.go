package template

func MaintainersMdTemplateGenerator() string {
	return `# Maintainers List

For notes on how to read this information, please refer to ` + "[`MAINTAINERS.md`]" + `(https://github.com/unikraft/unikraft/tree/staging/MAINTAINERS.md) in
the main Unikraft repository.

	{{ .LibKName }}-UNIKRAFT
	M:	{{ .AuthorName }} <{{ .AuthorEmail }}>
	L:	minios-devel@lists.xen.org
	F: *`
}
