package config

import (
	"bytes"
	"embed"
	"io"
	"reflect"
	"strings"
	"text/template"

	"github.com/ViaQ/logerr/v2/kverrors"
)

const (
	// LokiConfigFileName is the name of the config file in the configmap
	LokiConfigFileName = "config.yaml"
	// LokiRuntimeConfigFileName is the name of the runtime config file in the configmap
	LokiRuntimeConfigFileName = "runtime-config.yaml"
	// LokiConfigMountDir is the path that is mounted from the configmap
	LokiConfigMountDir = "/etc/loki/config"
)

var (
	//go:embed loki-config.yaml
	lokiConfigYAMLTmplFile embed.FS

	//go:embed loki-runtime-config.yaml
	lokiRuntimeConfigYAMLTmplFile embed.FS

	lokiConfigYAMLTmpl = template.Must(template.ParseFS(lokiConfigYAMLTmplFile, "loki-config.yaml"))

	lokiRuntimeConfigYAMLTmpl = template.Must(template.New("loki-runtime-config.yaml").Funcs(template.FuncMap{
		"yamlBlock": yamlBlock,
	}).ParseFS(lokiRuntimeConfigYAMLTmplFile, "loki-runtime-config.yaml"))
)

// Build builds a loki stack configuration files
func Build(lokiCustomConfigYAMLTmplStr []byte, opts Options) ([]byte, []byte, error) {
	var configYAMLTmpl *template.Template
	if len(lokiCustomConfigYAMLTmplStr) > 0 {
		tmpl, err := template.New("loki-config.yaml").Parse(string(lokiCustomConfigYAMLTmplStr))
		if err != nil {
			return nil, nil, kverrors.Wrap(err, "failed to create loki configuration YAML template")
		}
		configYAMLTmpl = tmpl
	} else {
		configYAMLTmpl = lokiConfigYAMLTmpl
	}
	// Build loki config yaml
	w := bytes.NewBuffer(nil)
	err := configYAMLTmpl.Execute(w, opts)
	if err != nil {
		return nil, nil, kverrors.Wrap(err, "failed to create loki configuration")
	}
	cfg, err := io.ReadAll(w)
	if err != nil {
		return nil, nil, kverrors.Wrap(err, "failed to read configuration from buffer")
	}
	// Build loki runtime config yaml
	w = bytes.NewBuffer(nil)
	err = lokiRuntimeConfigYAMLTmpl.Execute(w, opts)
	if err != nil {
		return nil, nil, kverrors.Wrap(err, "failed to create loki runtime configuration")
	}
	rcfg, err := io.ReadAll(w)
	if err != nil {
		return nil, nil, kverrors.Wrap(err, "failed to read configuration from buffer")
	}
	return cfg, rcfg, nil
}

// BuildWithTmpl builds a loki stack configuration files with custom template
func BuildWithTmpl(lokiConfigTmpl string, opts Options) ([]byte, []byte, error) {
	// Build loki config yaml with custom template
	lokiCustomConfigYAMLTmpl, err := template.New("loki-config.yaml").Parse(lokiConfigTmpl)
	if err != nil {
		return nil, nil, kverrors.Wrap(err, "failed to create loki configuration YAML template")
	}
	w := bytes.NewBuffer(nil)
	err = lokiCustomConfigYAMLTmpl.Execute(w, opts)
	if err != nil {
		return nil, nil, kverrors.Wrap(err, "failed to create loki configuration")
	}
	cfg, err := io.ReadAll(w)
	if err != nil {
		return nil, nil, kverrors.Wrap(err, "failed to read configuration from buffer")
	}
	// Build loki runtime config yaml
	w = bytes.NewBuffer(nil)
	err = lokiRuntimeConfigYAMLTmpl.Execute(w, opts)
	if err != nil {
		return nil, nil, kverrors.Wrap(err, "failed to create loki runtime configuration")
	}
	rcfg, err := io.ReadAll(w)
	if err != nil {
		return nil, nil, kverrors.Wrap(err, "failed to read configuration from buffer")
	}
	return cfg, rcfg, nil
}

func yamlBlock(indent string, in reflect.Value) string {
	inStr := in.String()
	lines := strings.Split(strings.TrimRight(inStr, "\n"), "\n")
	return strings.Join(lines, "\n"+indent)
}
