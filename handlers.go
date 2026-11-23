package main

import (
	"embed"
	"encoding/json"
	"html"
	"net/http"
)

type httpHandler struct {
	state    *applicationState
	webFiles embed.FS
}

func newHTTPHandler(applicationState *applicationState, webFiles embed.FS) http.Handler {
	handler := &httpHandler{state: applicationState, webFiles: webFiles}
	mux := http.NewServeMux()
	mux.HandleFunc("/", handler.serveWebPage)
	mux.HandleFunc("/obs.txt", handler.serveOBSText)
	mux.HandleFunc("/obs.html", handler.serveOBSOverlay)
	mux.HandleFunc("/deaths.txt", handler.serveDeathsText)
	mux.HandleFunc("/deaths.html", handler.serveDeathsOverlay)
	mux.HandleFunc("/api/state", handler.serveState)
	mux.HandleFunc("/api/settings", handler.saveSettings)
	mux.HandleFunc("/api/deaths", handler.updateDeaths)
	return mux
}

func (h *httpHandler) serveDeathsText(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = w.Write([]byte(h.state.deathText()))
}

func (h *httpHandler) serveDeathsOverlay(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	page := `<!doctype html><html><meta charset="utf-8"><style>html,body{margin:0;background:transparent;overflow:hidden}body{font-family:Arial,sans-serif;font-size:120px;font-weight:700;color:white;text-shadow:2px 2px 4px #000;white-space:nowrap}</style><body><span id="counter">` + html.EscapeString(h.state.deathText()) + `</span><script>setInterval(async()=>{try{const s=await(await fetch('/api/state')).json();document.getElementById('counter').textContent='Смерти: '+s.deaths}catch{}},1000)</script></body></html>`
	_, _ = w.Write([]byte(page))
}

func (h *httpHandler) updateDeaths(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST required", http.StatusMethodNotAllowed)
		return
	}
	var update struct {
		Deaths int `json:"deaths"`
	}
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if update.Deaths < 0 {
		http.Error(w, "invalid deaths count", http.StatusBadRequest)
		return
	}

	h.state.mu.Lock()
	h.state.Deaths = update.Deaths
	h.state.mu.Unlock()

	if err := saveApplicationState(h.state); err != nil {
		http.Error(w, "could not save deaths", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *httpHandler) serveOBSOverlay(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
	w.Header().Set("Pragma", "no-cache")
	text := html.EscapeString(h.state.obsText())
	page := "<!doctype html><html><meta charset=\"utf-8\"><style>html,body{margin:0;background:transparent;overflow:hidden}body{font-family:Arial,sans-serif;font-size:120px;font-weight:700;color:white;text-shadow:2px 2px 4px #000;white-space:nowrap}</style><body><span id=\"counter\">" + text + "</span><script>setInterval(async()=>{try{const state=await (await fetch('/api/state')).json();document.getElementById('counter').textContent=(state.label?state.label+': ':'')+state.killed+'/'+state.total}catch{}},1000)</script></body></html>"
	_, _ = w.Write([]byte(page))
}

func (h *httpHandler) serveWebPage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		data, err := h.webFiles.ReadFile("web/index.html")
		if err != nil {
			http.Error(w, "interface file is unavailable", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(data)
		return
	}

	fileServer := http.FileServer(http.FS(h.webFiles))
	fileServer.ServeHTTP(w, r)
}

func (h *httpHandler) serveOBSText(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = w.Write([]byte(h.state.obsText()))
}

func (h *httpHandler) serveState(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err := json.NewEncoder(w).Encode(h.state.snapshot()); err != nil {
		http.Error(w, "could not encode state", http.StatusInternalServerError)
	}
}

func (h *httpHandler) saveSettings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST required", http.StatusMethodNotAllowed)
		return
	}

	var settings struct {
		Label string `json:"label"`
	}
	if err := json.NewDecoder(r.Body).Decode(&settings); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	// Update the shared settings while holding the state lock.
	h.state.mu.Lock()
	h.state.Label = settings.Label
	h.state.mu.Unlock()

	if err := saveApplicationState(h.state); err != nil {
		http.Error(w, "could not save settings", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
