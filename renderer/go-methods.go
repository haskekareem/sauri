package renderer

import (
	"bytes"
	"fmt"
	"github.com/justinas/nosurf"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
)

// todo: Go template engine support

// AddCustomFuncs adds a custom template function to the Render instance.
func (r *Renderer) AddCustomFuncs(customFuncs template.FuncMap) {
	if r.CustomFuncs == nil {
		r.CustomFuncs = make(template.FuncMap)
	}

	for name, fn := range customFuncs {
		r.CustomFuncs[name] = fn
	}
}

// ParseTemplates parses all templates in the directory and cache as map.
func (r *Renderer) ParseTemplates() error {
	// layouts template
	layoutFiles, err := filepath.Glob(filepath.Join(r.TemplatesRootPath, "views", "layouts", "*layout.gohtml"))
	if err != nil {
		return fmt.Errorf("error globbing layout files: %v", err)
	}
	// Page template
	Pages, err := filepath.Glob(filepath.Join(r.TemplatesRootPath, "views", "pages", "*.gohtml"))
	if err != nil {
		return fmt.Errorf("error globbing pages files: %v", err)
	}

	for _, page := range Pages {
		files := append(layoutFiles, page)
		name := filepath.Base(page)
		tmpl, err := template.New(name).Funcs(r.CustomFuncs).ParseFiles(files...)
		if err != nil {
			return fmt.Errorf("error parsing template %s: %v", name, err)
		}
		r.GoTemplateCache.Store(name, tmpl)
	}
	return nil
}

// cacheTemplates ensures templates are cached once in production mode.
func (r *Renderer) cacheTemplates() {
	// Ensures the function inside is executed only once
	r.once.Do(func() {
		if err := r.ParseTemplates(); err != nil {
			log.Printf("Failed to load and cache templates: %v\n", err)
		}

	})
}

// AddDefaultsData add common dynamic data on every webpage
func (r *Renderer) AddDefaultsData(td *TemplateData, rr *http.Request) *TemplateData {
	if td == nil {
		td = r.NewTemplateData()
	}

	td.ServerName = r.ServeName
	td.CSRFToken = nosurf.Token(rr)
	td.Port = r.Port
	td.Secure = r.Secure

	if r.Session.Exists(rr.Context(), "userID") {
		td.IsUserAuthenticated = true
	}

	return td
}

// getTemplate retrieves the specified template from the cache or loads it if in development mode.
func (r *Renderer) getTemplate(tempName string) (*template.Template, error) {
	if r.DevelopmentMode {
		// Reload templates on each request in development mode
		if err := r.ParseTemplates(); err != nil {
			log.Printf("error parsing templates: %v\n", err)
			return nil, err
		}
	} else {
		// Ensure templates are cached only once in production mode
		r.cacheTemplates()
	}

	// Retrieve the template from the cache stored as a Map
	tmp, ok := r.GoTemplateCache.Load(tempName)
	if !ok {
		return nil, fmt.Errorf("template %s does not exist", tempName)
	}

	return tmp.(*template.Template), nil

}

// RenderGoPage retrieves the specified template from the cache or loads it
// if in development mode and then executes it.
func (r *Renderer) RenderGoPage(w http.ResponseWriter, rr *http.Request, tmpl string, data any) error {
	// retrieve the specified template
	tmp, err := r.getTemplate(tmpl)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	// check if data is passed to the template
	var td *TemplateData
	if data != nil {
		var ok bool
		td, ok = data.(*TemplateData)
		if !ok {
			http.Error(w, "Invalid template data.", http.StatusInternalServerError)
			return nil
		}
	}

	// pass in some default data to all templates
	td = r.AddDefaultsData(td, rr)

	// Execute the template
	buf := new(bytes.Buffer)
	if err := tmp.Execute(buf, td); err != nil {
		log.Printf("error executing template to buffer: %v\n", err)
		http.Error(w, "Error buffer template.", http.StatusInternalServerError)
		return err
	}

	// write the content to the web browser
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-XSS-Protection", "1; mode=block")
	w.Header().Set("X-Frame-Options", "deny")

	if _, err := buf.WriteTo(w); err != nil {
		log.Printf("error writing template to the browser: %v\n", err)
		http.Error(w, "Error rendering template.", http.StatusInternalServerError)
		return err
	}

	return nil
}
