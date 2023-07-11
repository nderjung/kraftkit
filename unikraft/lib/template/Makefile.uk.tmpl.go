package template

func MakefileUkGenerator() string {
	return `# SPDX-License-Identifier: BSD-3-Clause
#
{{if .Description }}
# {{ .Description }}
{{else}}
# {{ .ProjectName }} Unikraft library
{{ end }}
#
# Authors: {{ .AuthorName }} <{{ .AuthorEmail }}>
#
# Copyright (c) {{ .Year }}, {{ .CopyrightHolder }}. All rights reserved.
#
# Redistribution and use in source and binary forms, with or without
# modification, are permitted provided that the following conditions
# are met:
#
# 1. Redistributions of source code must retain the above copyright
#    notice, this list of conditions and the following disclaimer.
# 2. Redistributions in binary form must reproduce the above copyright
#    notice, this list of conditions and the following disclaimer in the
#    documentation and/or other materials provided with the distribution.
# 3. Neither the name of the copyright holder nor the names of its
#    contributors may be used to endorse or promote products derived from
#    this software without specific prior written permission.
#
# THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
# AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
# IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
# ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE
# LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
# CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
# SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
# INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
# CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
# ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
# POSSIBILITY OF SUCH DAMAGE.

################################################################################
# Library registration
################################################################################
$(eval $(call addlib_s,{{ .LibKName }},$(CONFIG_{{ .LibKName }})))

################################################################################
# Sources
################################################################################]
{{if .Commit}}
{{ .LibKName }}_COMMIT = {{ .Commit }}
{{ end }}
{{ if .Version }}
{{ .LibKName|upper }}_VERSION = {{ .Version }}
{{ end }}
{{ .LibKName|upper }}_URL = {{ .Origin_url }}
{{ .LibKName|upper }}_SUBDIR = {{ .ProjectName }}-$({{ .LibKName|upper }}_VERSION)
{{if .WithPatchedir }}
{{ .LibKName|upper }}_PATCHDIR = $({{ .LibKName|upper }}_BASE)/patches
$(eval $(call patch,{{ .LibKName }},$({{ .LibKName|upper }}_PATCHDIR),$({{ .LibKName|upper }}_SUBDIR)))
{{ end }}
$(eval $(call fetch,{{ .LibKName|lower }},$({{ .LibKName|upper }}_URL)))

################################################################################
# Helpers
################################################################################
{{ .LibKName|upper }}_SRC = $({{ .LibKName|upper }}_ORIGIN)/$({{ .LibKName|upper }}_SUBDIR)

################################################################################
# Library includes
################################################################################
CINCLUDES-y += -I$({{ .LibKName|upper }}_BASE)/include

################################################################################
# Flags
################################################################################
{{ .LibKName|upper }}_FLAGS =

# Suppress some warnings to make the build process look neater
{{ .LibKName|upper }}_FLAGS_SUPPRESS =

{{ .LibKName|upper }}_CFLAGS-y += $({{ .LibKName|upper }}_FLAGS)
{{ .LibKName|upper }}_CFLAGS-y += $({{ .LibKName|upper }}_FLAGS_SUPPRESS)

################################################################################
# Glue code
################################################################################
# Include paths
# {{ .LibKName|upper }}_CINCLUDES-y   += $({{ .LibKName|upper }}_COMMON_INCLUDES-y)
# {{ .LibKName|upper }}_CXXINCLUDES-y += $({{ .LibKName|upper }}_COMMON_INCLUDES-y)

{{ if .ProvideMain }}
{{ .LibKName|upper }}SRCS-$(CONFIG_{{ .LibKName|upper }}_MAIN_FUNCTION) += $({{ .LibKName|upper }}_BASE)/main.c|unikraft
{{end}}

################################################################################
# Library sources
################################################################################
# {{ .LibKName|upper }}_SRCS-y += # Include source files here

{{range $index, $source_file := .Source_files }}
{{ .LibKName|upper }}_SRCS-y += {{ $source_file }}
{{end}}`
}
