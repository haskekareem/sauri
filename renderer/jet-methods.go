package renderer

import (
	"github.com/CloudyKit/jet/v6"
	"net/http"
	"path"
	"strings"
)

// todo: Jet template engine support

// RenderJetPage renders a template using Jet Template engine
func (r *Renderer) RenderJetPage(w http.ResponseWriter, rr *http.Request, temName string, variable, data any) error {
	// 1) Normalize and sanitize the template name:
	cleanName := strings.Trim(path.Clean(temName), "/")
	tplPath := cleanName + ".jet"

	// 2) Prepare Jet variables map:
	var vars jet.VarMap
	if variable == nil {
		vars = make(jet.VarMap)
	} else {
		vars = variable.(jet.VarMap)
	}

	// 3) Prepare template data context:
	td := &TemplateData{}
	if data != nil {
		td = data.(*TemplateData)
	}

	td = r.AddDefaultsData(td, rr)

	// retrieving the specified template to be display
	t, err := r.JetViews.GetTemplate(tplPath)
	if err != nil {
		//log.Printf("Error loading template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return err
	}

	// execute the template to the web browser
	if err = t.Execute(w, vars, td); err != nil {
		//log.Printf("Error executing template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return err
	}

	return nil
}
