package renderer

import (
	"github.com/CloudyKit/jet/v6"
	"github.com/alexedwards/scs/v2"
	"html/template"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

// Renderer struct to hold templates and custom functions
type Renderer struct {
	RendererEngine    string
	TemplatesRootPath string
	Secure            bool
	Port              string
	ServeName         string
	GoTemplateCache   sync.Map
	JetViews          *jet.Set
	once              sync.Once
	CustomFuncs       template.FuncMap
	DefaultData       *TemplateData
	DevelopmentMode   bool
	Session           *scs.SessionManager
}

type TemplateData struct {
	IsUserAuthenticated bool
	IntMap              map[string]int
	StringMap           map[string]string
	FloatMap            map[string]float64
	GenericData         map[string]any
	CSRFToken           string //for CSRF Protection implementation
	Secure              bool
	Port                string
	ServerName          string
	FormData            url.Values
	Errors              map[string][]string
}

// NewTemplateData returns a new instance of TemplateData with all maps initialized.
func (r *Renderer) NewTemplateData() *TemplateData {
	return &TemplateData{
		IsUserAuthenticated: false,
		IntMap:              make(map[string]int),
		StringMap:           make(map[string]string),
		FloatMap:            make(map[string]float64),
		GenericData:         make(map[string]any),
		CSRFToken:           "",
		Secure:              false,
		Port:                "",
		ServerName:          "",
		FormData:            nil,
		Errors:              nil,
	}
}

// RenderPage specifies default template rendering engine
func (r *Renderer) RenderPage(w http.ResponseWriter, rr *http.Request, temName string, variable, data any) error {
	switch strings.ToLower(r.RendererEngine) {
	case "go":
		return r.RenderGoPage(w, rr, temName, data)
	case "jet":
		return r.RenderJetPage(w, rr, temName, variable, data)
	}
	return nil
}
