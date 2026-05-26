package api

import "net/http"

func (s *Server) pageSetup(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	s.render(w, "simulation_setup.html", nil)
}

func (s *Server) pageLive(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	s.mu.Lock()
	st := s.sims[id]
	s.mu.Unlock()
	name := ""
	if st != nil {
		name = st.req.Name
	}
	s.render(w, "live_view.html", map[string]any{"ID": id, "Name": name})
}

func (s *Server) pageReport(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	http.Redirect(w, r, "/reports/"+id+".html", http.StatusFound)
}

func (s *Server) render(w http.ResponseWriter, name string, data any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.tmpl.ExecuteTemplate(w, name, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
