package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewTemplateResolver(t *testing.T) {
	tr := NewTemplateResolver("myapp", "/project", "/project", "/cwd")
	if tr == nil {
		t.Fatal("NewTemplateResolver() returned nil")
	}
	if tr.context["APP_NAME"] != "myapp" {
		t.Errorf("APP_NAME = %v, want myapp", tr.context["APP_NAME"])
	}
	if tr.context["PROJECT_ROOT"] != "/project" {
		t.Errorf("PROJECT_ROOT = %v, want /project", tr.context["PROJECT_ROOT"])
	}
	if tr.context["WORKDIR"] != "/project" {
		t.Errorf("WORKDIR = %v, want /project", tr.context["WORKDIR"])
	}
	if tr.context["CWD"] != "/cwd" {
		t.Errorf("CWD = %v, want /cwd", tr.context["CWD"])
	}
}

func TestResolveString(t *testing.T) {
	tr := NewTemplateResolver("myapp", "/project", "/project", "/cwd")

	tests := []struct {
		name string
		in   string
		want string
	}{
		{"empty", "", ""},
		{"no vars", "hello world", "hello world"},
		{"single var", "${APP_NAME}", "myapp"},
		{"multiple vars", "${APP_NAME}-${WORKDIR}", "myapp-/project"},
		{"unknown var", "${UNKNOWN_VAR}", "${UNKNOWN_VAR}"},
		{"mixed", "app=${APP_NAME} dir=${WORKDIR}", "app=myapp dir=/project"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tr.ResolveString(tt.in)
			if got != tt.want {
				t.Errorf("ResolveString(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestResolveStringMap(t *testing.T) {
	tr := NewTemplateResolver("myapp", "/project", "/project", "/cwd")

	if tr.ResolveStringMap(nil) != nil {
		t.Error("ResolveStringMap(nil) should return nil")
	}

	m := map[string]string{"a": "${APP_NAME}", "b": "literal"}
	got := tr.ResolveStringMap(m)
	if got["a"] != "myapp" {
		t.Errorf("ResolveStringMap a = %v, want myapp", got["a"])
	}
	if got["b"] != "literal" {
		t.Errorf("ResolveStringMap b = %v, want literal", got["b"])
	}
}

func TestResolveStringSlice(t *testing.T) {
	tr := NewTemplateResolver("myapp", "/project", "/project", "/cwd")

	if tr.ResolveStringSlice(nil) != nil {
		t.Error("ResolveStringSlice(nil) should return nil")
	}

	s := []string{"${APP_NAME}", "static"}
	got := tr.ResolveStringSlice(s)
	if len(got) != 2 {
		t.Fatalf("ResolveStringSlice len = %v, want 2", len(got))
	}
	if got[0] != "myapp" {
		t.Errorf("ResolveStringSlice[0] = %v, want myapp", got[0])
	}
	if got[1] != "static" {
		t.Errorf("ResolveStringSlice[1] = %v, want static", got[1])
	}
}

func TestResolveService(t *testing.T) {
	tr := NewTemplateResolver("myapp", "/project", "/project", "/cwd")

	if tr.ResolveService(nil) != nil {
		t.Error("ResolveService(nil) should return nil")
	}

	svc := &Service{
		Name:    "svc1",
		Command: "echo ${APP_NAME}",
		WorkDir: "${WORKDIR}",
		Port:    8080,
	}
	got := tr.ResolveService(svc)
	if got == nil {
		t.Fatal("ResolveService() returned nil")
	}
	if got.Name != "svc1" {
		t.Errorf("Name = %v, want svc1", got.Name)
	}
	if got.Command != "echo myapp" {
		t.Errorf("Command = %v, want 'echo myapp'", got.Command)
	}
	// Resolved /project is normalized to . when same as app WORKDIR
	if got.WorkDir != "." && got.WorkDir != "/project" {
		t.Errorf("WorkDir = %v, want . or /project", got.WorkDir)
	}
	if got.Port != 8080 {
		t.Errorf("Port = %v, want 8080", got.Port)
	}
}

func TestResolveService_Docker(t *testing.T) {
	tr := NewTemplateResolver("myapp", "/project", "/project", "/cwd")
	svc := &Service{
		Name: "db",
		Type: "docker",
		Docker: &DockerConfig{
			Image:   "img:${APP_NAME}",
			Container: "cnt-${WORKDIR}",
			Ports:   []string{"5432:5432"},
			Volumes: []string{"${WORKDIR}/data:/data"},
			Environment: map[string]string{"X": "${APP_NAME}"},
		},
	}
	got := tr.ResolveService(svc)
	if got == nil || got.Docker == nil {
		t.Fatal("ResolveService() returned nil or Docker nil")
	}
	if got.Docker.Image != "img:myapp" {
		t.Errorf("Docker.Image = %v, want img:myapp", got.Docker.Image)
	}
	if got.Docker.Environment["X"] != "myapp" {
		t.Errorf("Docker.Environment[X] = %v, want myapp", got.Docker.Environment["X"])
	}
}

func TestResolveAppSpec(t *testing.T) {
	tr := NewTemplateResolver("myapp", "/project", "/project", "/cwd")

	if tr.ResolveAppSpec(nil) != nil {
		t.Error("ResolveAppSpec(nil) should return nil")
	}

	app := &AppSpec{
		Name:        "myapp",
		WorkDir:     "${WORKDIR}",
		Services:    []Service{{Name: "s1", Command: "echo ${APP_NAME}"}},
		Modes:       map[string]Mode{"default": {Services: []string{"s1"}}},
		Environment: map[string]string{"K": "${APP_NAME}"},
	}
	got := tr.ResolveAppSpec(app)
	if got == nil {
		t.Fatal("ResolveAppSpec() returned nil")
	}
	if got.WorkDir != "/project" {
		t.Errorf("WorkDir = %v, want /project", got.WorkDir)
	}
	if len(got.Services) != 1 || got.Services[0].Command != "echo myapp" {
		t.Errorf("Services[0].Command = %v, want 'echo myapp'", got.Services[0].Command)
	}
	if got.Environment["K"] != "myapp" {
		t.Errorf("Environment[K] = %v, want myapp", got.Environment["K"])
	}
}

func TestResolveAppSpec_Overrides(t *testing.T) {
	tr := NewTemplateResolver("myapp", "/project", "/project", "/cwd")
	app := &AppSpec{
		Name:    "myapp",
		WorkDir: ".",
		Services: []Service{{Name: "s1", Command: "cmd"}},
		Modes: map[string]Mode{
			"default": {
				Services: []string{"s1"},
				Environment: map[string]string{"M": "mval"},
				Overrides: []ServiceOverride{
					{ServiceName: "s1", Command: "override-${APP_NAME}"},
				},
			},
		},
	}
	got := tr.ResolveAppSpec(app)
	if got == nil {
		t.Fatal("ResolveAppSpec() returned nil")
	}
	m := got.Modes["default"]
	if len(m.Overrides) != 1 || m.Overrides[0].Command != "override-myapp" {
		t.Errorf("Overrides[0].Command = %v, want override-myapp", m.Overrides[0].Command)
	}
}

func TestResolveFilePath(t *testing.T) {
	home, _ := os.UserHomeDir()
	if home == "" {
		home = os.Getenv("HOME")
	}
	tr := NewTemplateResolver("app", "/p", "/p", "/cwd")

	got := tr.ResolveFilePath("/abs/path")
	if got != "/abs/path" {
		t.Errorf("ResolveFilePath(/abs/path) = %q, want /abs/path", got)
	}
	got = tr.ResolveFilePath("${APP_NAME}")
	if got != "app" {
		t.Errorf("ResolveFilePath(${APP_NAME}) = %q, want app", got)
	}
	if home != "" {
		got = tr.ResolveFilePath("~/foo")
		want := filepath.Join(home, "foo")
		if got != want {
			t.Errorf("ResolveFilePath(~/foo) = %q, want %q", got, want)
		}
	}
}
