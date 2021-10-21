package views

import (
	"html/template"
	"net/http"
	"path/filepath"
)

var (
	LayoutDir   string = "views/layouts/"
	TemplateDir string = "views/"
	TemplateExt string = ".gohtml"
)

func NewView(layout string, files ...string) *View {
	addTemplatePath(files)
	addTemplateExt(files)
	files = append(files, layoutFiles()...)
	t, err := template.ParseFiles(files...)
	if err != nil {
		panic(err)
	}

	return &View{
		Template: t,
		Layout:   layout,
	}
}

type View struct {
	Template *template.Template
	Layout   string
}

// Implements the Hanlder interface. Renders a View object, and
// allows a static view to be rendered directly from Handle()
func (v *View) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := v.Render(w, nil); err != nil {
		panic(err)
	}
}

// Renders the view using the predefined layout
func (v *View) Render(w http.ResponseWriter, data interface{}) error {
	w.Header().Set("Content-Type", "text/html")
	return v.Template.ExecuteTemplate(w, v.Layout, data)
}

// Returns a slice of strings representing layout
// files used in the application.
func layoutFiles() []string {
	files, err := filepath.Glob(LayoutDir + "*" + TemplateExt)
	if err != nil {
		panic(err)
	}

	return files
}

// Takes a slice of strings representing relative file paths for templates,
// and prepends the template directory path to each template file path.
// E.g. input {"home"} results in {"views/home"} if TemplateDir is "views/"
func addTemplatePath(files []string) {
	for i, f := range files {
		files[i] = TemplateDir + f
	}
}

// Takes a slice of strings representing relative file paths for templates,
// and appends the TemplateExt extension to each string in the slice.
// E.g. input {"home"} results in {"home.gohtml"} if TemplateExt is ".gohtml"
func addTemplateExt(files []string) {
	for i, f := range files {
		files[i] = f + TemplateExt
	}
}
