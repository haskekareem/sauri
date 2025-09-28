package renderer

import (
	"fmt"
	"github.com/CloudyKit/jet/v6"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"html/template"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// --- SETUP HELPERS ---

func setTestRenderer(engine string, devMode bool, root string) *Renderer {
	r := &Renderer{
		RendererEngine:    engine,
		Secure:            false,
		Port:              "8080",
		TemplatesRootPath: root,
		ServeName:         "testServer",
		//CustomFuncs:       template.FuncMap{},
		//DefaultData:       &TemplateData{},
		DevelopmentMode: devMode,
		//GoTemplateCache: sync.Map{},
		JetViews: jet.NewSet(
			jet.NewOSFileSystemLoader(filepath.Join("resources-test", "views")),
			jet.InDevelopmentMode()),
	}
	return r
}

func writeGoTemplates(t *testing.T, root, layName, temName string) {
	// create the template files
	layoutDir := filepath.Join(root, "views", "layouts")
	pageDir := filepath.Join(root, "views", "pages")

	layoutContent := `
	{{define "base"}}
    <!DOCTYPE html>
    <html lang="en">

    <head>
        <meta charset="UTF-8">
        <meta http-equiv="X-UA-Compatible" content="IE=edge">
        <meta name="viewport" content="width=device-width, initial-scale=1.0">
        <title>My page</title>
    </head>

    <body>

    <!-- *define a section*-->
    <!--* it tell golang that the content in this block will change per template basis *-->
    {{block "content" .}}


    {{end}}


    <!-- js code specific for different pages-->
    {{block "js" .}}


    {{end}}


    </body>

    </html>

{{end}}
`
	pageContent := `
		{{template "base" .}} <!-- using the base layout template-->
		
		{{define "content"}}
		
			<h1>this is the home page</h1>
		
		{{end}}`

	require.NoError(t, os.WriteFile(filepath.Join(layoutDir, layName), []byte(layoutContent), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(pageDir, temName), []byte(pageContent), 0644))

}

func writeGoTemplatesWithFuncs(t *testing.T, root, layName, temName string) {
	// create the template files
	layoutDir := filepath.Join(root, "views", "layouts")
	pageDir := filepath.Join(root, "views", "pages")

	layoutContent := `
	{{define "base"}}
    <!DOCTYPE html>
    <html lang="en">

    <head>
        <meta charset="UTF-8">
        <meta http-equiv="X-UA-Compatible" content="IE=edge">
        <meta name="viewport" content="width=device-width, initial-scale=1.0">
        <title>My page</title>
    </head>

    <body>

    <!-- *define a section*-->
    <!--* it tell golang that the content in this block will change per template basis *-->
    {{block "content" .}}


    {{end}}


    <!-- js code specific for different pages-->
    {{block "js" .}}


    {{end}}


    </body>

    </html>

{{end}}
`
	pageContent := `
		{{template "base" .}} <!-- using the base layout template-->
		
		{{define "content"}}
		
			<h1>my name is  {{ ToUpper "nurudeen" }}</h1>
		
		{{end}}`

	require.NoError(t, os.WriteFile(filepath.Join(layoutDir, layName), []byte(layoutContent), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(pageDir, temName), []byte(pageContent), 0644))

}

func writeJetTemplate(t *testing.T, root string, temName string) {
	viewsDir := filepath.Join(root, "views")

	tpl := `Hello from Jet {{ Title }}`
	require.NoError(t, os.WriteFile(filepath.Join(viewsDir, fmt.Sprintf("%s.jet", temName)), []byte(tpl), 0644))
}

// --- TESTS ---

// Test_RenderGoPage_Success tests successful rendering using Go templates
func Test_RenderGoPage_Success(t *testing.T) {

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp := httptest.NewRecorder()

	writeGoTemplates(t, "resources-test", "base.layout.gohtml", "home.page.gohtml")

	// Now initialize renderer (it will read the newly written templates)
	r := setTestRenderer("go", true, "resources-test")

	// Render a known template
	err := r.RenderGoPage(resp, req, "home.page.gohtml", nil)
	require.NoError(t, err)

	// Check for HTML response
	body := resp.Body.String()
	assert.Contains(t, body, "this is the home page")

	layoutDir := filepath.Join("resources-test", "views", "layouts")
	pageDir := filepath.Join("resources-test", "views", "pages")

	defer os.Remove(filepath.Join(layoutDir, "base.layout.gohtml"))
	defer os.Remove(filepath.Join(pageDir, "home.page.gohtml"))

}

// Test_RenderGoPage_MissingTemplate tests rendering failure due to missing template.
func Test_RenderGoPage_MissingTemplate(t *testing.T) {

	writeGoTemplates(t, "resources-test", "base.layout.gohtml", "index.page.gohtml")

	r := setTestRenderer("go", false, "resources-test")

	resp := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)

	// Template does not exist in cache
	err := r.RenderGoPage(resp, req, "nonexistent.page.gohtml", nil)
	if err == nil {
		t.Error("Expected error for missing template, got nil")
	}

	layoutDir := filepath.Join("resources-test", "views", "layouts")
	pageDir := filepath.Join("resources-test", "views", "pages")

	defer os.Remove(filepath.Join(layoutDir, "base.layout.gohtml"))
	defer os.Remove(filepath.Join(pageDir, "index.page.gohtml"))

}

// Test_RenderJetPage_Success tests Jet template rendering and also check its content.
func Test_RenderJetPage_Success(t *testing.T) {

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)

	writeJetTemplate(t, "resources-test", "index")

	r := setTestRenderer("jet", true, "resources-test")

	// Create jet variables
	vars := make(jet.VarMap)
	vars.Set("Title", "Welcome")

	// Render an existing Jet template
	err := r.RenderJetPage(w, req, "index", vars, nil)
	require.NoError(t, err)

	body := w.Body.String()
	assert.Contains(t, body, "Hello from Jet Welcome")

	pageDir := filepath.Join("resources-test", "views")
	defer os.Remove(filepath.Join(pageDir, "index.jet"))

}

// Test_RenderJetPage_MissingTemplate tests rendering with a missing Jet template.
func Test_RenderJetPage_MissingTemplate(t *testing.T) {
	writeJetTemplate(t, "resources-test", "index")

	r := setTestRenderer("jet", false, "resources-test")

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)

	// attempt to render non-existent template
	err := r.RenderJetPage(w, req, "non-exist", nil, nil)
	if err == nil {
		t.Error("Expected error for missing template, got nil")
	}

	pageDir := filepath.Join("resources-test", "views")
	defer os.Remove(filepath.Join(pageDir, "index.jet"))
}

// Test_RenderPage_Router tests the main routing logic of RenderPage() function.
func Test_RenderPage_Router(t *testing.T) {
	writeGoTemplates(t, "resources-test", "base.layout.gohtml", "about.page.gohtml")
	writeJetTemplate(t, "resources-test", "contact")

	tests := []struct {
		name            string
		engine          string
		templateName    string
		developmentMode bool
		expectedError   bool
	}{
		{"Valid Go Template", "go", "about.page.gohtml", true, false},
		{"Invalid Go Template", "go", "non-exsit.gohtml", true, true},
		{"Valid Jet Template", "jet", "contact", true, false},
		{"Invalid Jet Template", "jet", "non-exsit", true, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/", nil)

			r := setTestRenderer(test.engine, test.developmentMode, "resources-test")

			var vars jet.VarMap
			if test.engine == "jet" && test.templateName == "contact" {
				vars = make(jet.VarMap)
				vars.Set("Title", "Welcome")
			}

			err := r.RenderPage(w, req, test.templateName, vars, nil)

			if test.expectedError {
				//require.Error(t, err)
				if err == nil {
					t.Error("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error, got %v", err)
				}
				//require.NoError(t, err)
			}
		})
	}

	layoutDir := filepath.Join("resources-test", "views", "layouts")
	pageDir := filepath.Join("resources-test", "views", "pages")
	jetDir := filepath.Join("resources-test", "views")

	defer os.Remove(filepath.Join(layoutDir, "base.layout.gohtml"))
	defer os.Remove(filepath.Join(pageDir, "about.page.gohtml"))
	defer os.Remove(filepath.Join(jetDir, "contact.jet"))
}

// Test_RenderGoPage_InvalidDataType tests rendering with an invalid data type.
func Test_RenderGoPage_InvalidDataType(t *testing.T) {
	writeGoTemplates(t, "resources-test", "base.layout.gohtml", "index.page.gohtml")

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)

	r := setTestRenderer("go", true, "resources-test")

	invalid := make(chan int)

	err := r.RenderPage(w, req, "index", nil, invalid)
	if err == nil {
		t.Error("Expected error, got nil")
	}

	layoutDir := filepath.Join("resources-test", "views", "layouts")
	pageDir := filepath.Join("resources-test", "views", "pages")

	defer os.Remove(filepath.Join(layoutDir, "base.layout.gohtml"))
	defer os.Remove(filepath.Join(pageDir, "index.page.gohtml"))
}

func Test_RenderGoPage_WithCustomFunction(t *testing.T) {
	writeGoTemplatesWithFuncs(t, "resources-test", "base.layout.gohtml", "customfunc.page.gohtml")

	r := setTestRenderer("go", true, "resources-test")

	// Register custom function
	r.AddCustomFuncs(template.FuncMap{
		"ToUpper": strings.ToUpper,
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	err := r.RenderPage(w, req, "customfunc.page.gohtml", nil, nil)
	if err != nil {
		t.Errorf("expected no error when rendering with custom function, got: %v", err)
	}

	got := w.Body.String()
	expected := "NURUDEEN"
	assert.Contains(t, got, expected)

	layoutDir := filepath.Join("resources-test", "views", "layouts")
	pageDir := filepath.Join("resources-test", "views", "pages")

	defer os.Remove(filepath.Join(layoutDir, "base.layout.gohtml"))
	defer os.Remove(filepath.Join(pageDir, "customfunc.page.gohtml"))
}

func writeGoTemplatesWithDefaultData(t *testing.T, root, layName, temName string) {
	layoutDir := filepath.Join(root, "views", "layouts")
	pageDir := filepath.Join(root, "views", "pages")

	layoutContent := `
	{{define "base"}}
	<!DOCTYPE html>
	<html>
	<body>
	{{block "content" .}}{{end}}
	</body>
	</html>
	{{end}}`

	pageContent := `
	{{template "base" .}}
	{{define "content"}}
	<p>Server: {{ .ServerName }}</p>
	<p>CustomData: {{ .StringMap.customKey }}</p>
	{{end}}`

	require.NoError(t, os.WriteFile(filepath.Join(layoutDir, layName), []byte(layoutContent), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(pageDir, temName), []byte(pageContent), 0644))
}

/*
func Test_RenderGoPage_WithDefaultData(t *testing.T) {
	writeGoTemplatesWithDefaultData(t, "resources-test", "base.layout.gohtml", "defaultdata.page.gohtml")

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	r := setTestRenderer("go", true, "resources-test")

	td := NewTemplateData()

	td.StringMap["customKey"] = "HelloFromCustomDefaults"

	dta := r.AddDefaultsData(td, req)

	err := r.RenderPage(w, req, "defaultdata.page.gohtml", nil, dta)
	require.NoError(t, err)

	got := w.Body.String()
	assert.Contains(t, got, "HelloFromCustomDefaults")
	assert.Contains(t, got, "testServer")

	layoutDir := filepath.Join("resources-test", "views", "layouts")
	pageDir := filepath.Join("resources-test", "views", "pages")

	defer os.Remove(filepath.Join(layoutDir, "base.layout.gohtml"))
	defer os.Remove(filepath.Join(pageDir, "defaultdata.page.gohtml"))
}
*/
