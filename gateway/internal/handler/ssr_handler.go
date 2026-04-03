package handler

import (
	"embed"
	"gateway/internal/client"
	"html/template"
	"log"
	"net/http"

	pb "shared/proto"

	"github.com/gorilla/mux"
)

//go:embed templates/*.html
var templateFS embed.FS

type SSRHandler struct {
	authClient  *client.AuthClient
	videoClient *client.VideoClient
	templates   map[string]*template.Template
}

func NewSSRHandler(authClient *client.AuthClient, videoClient *client.VideoClient) *SSRHandler {
	s := &SSRHandler{
		authClient:  authClient,
		videoClient: videoClient,
		templates:   make(map[string]*template.Template),
	}
	s.parseTemplates()
	return s
}

func (h *SSRHandler) parseTemplates() {
	layoutPath := "templates/layout.html"
	pages := []string{"login", "register", "home", "video_detail"}

	for _, page := range pages {
		pagePath := "templates/" + page + ".html"
		tmpl, err := template.ParseFS(templateFS, layoutPath, pagePath)
		if err != nil {
			log.Fatalf("Error parsing template %s: %v", page, err)
		}
		h.templates[page] = tmpl
	}
}

func (h *SSRHandler) ShowLogin(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"Title": "Login - TextTube",
	}
	if err := h.templates["login"].ExecuteTemplate(w, "layout.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *SSRHandler) ShowRegister(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"Title": "Register - TextTube",
	}
	if err := h.templates["register"].ExecuteTemplate(w, "layout.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *SSRHandler) Login(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	password := r.FormValue("password")

	resp, err := h.authClient.Login(r.Context(), &pb.LoginRequest{
		Email:    email,
		Password: password,
	})

	if err != nil {
		data := map[string]interface{}{
			"Title": "Login - TextTube",
			"Error": "Invalid email or password",
		}
		h.templates["login"].ExecuteTemplate(w, "layout.html", data)
		return
	}

	// Set HttpOnly cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    resp.Token,
		Path:     "/",
		HttpOnly: true,
		Secure:   true, // Should be true in prod
		SameSite: http.SameSiteStrictMode,
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *SSRHandler) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:   "session_token",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (h *SSRHandler) Register(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	email := r.FormValue("email")
	password := r.FormValue("password")

	resp, err := h.authClient.Register(r.Context(), &pb.RegisterRequest{
		Username: username,
		Email:    email,
		Password: password,
	})

	if err != nil {
		data := map[string]interface{}{
			"Title": "Register - TextTube",
			"Error": "Registration failed",
		}
		h.templates["register"].ExecuteTemplate(w, "layout.html", data)
		return
	}

	// Set HttpOnly cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    resp.Token,
		Path:     "/",
		HttpOnly: true,
		Secure:   true, // Should be true in prod
		SameSite: http.SameSiteStrictMode,
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *SSRHandler) Home(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	username := r.Context().Value("username").(string)
	query := r.URL.Query().Get("q")

	data := map[string]interface{}{
		"Title":         "Home - TextTube",
		"Authenticated": true,
		"Username":      username,
		"Query":         query,
	}

	if query != "" {
		resp, err := h.videoClient.SearchChannel(r.Context(), &pb.SearchChannelRequest{
			ChannelName: query,
			UserId:      userID,
		})
		if err == nil {
			data["Videos"] = resp.Videos
		} else {
			log.Printf("Search error: %v", err)
		}
	}

	if err := h.templates["home"].ExecuteTemplate(w, "layout.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *SSRHandler) VideoDetail(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	videoID := vars["videoId"]
	userID := r.Context().Value("user_id").(string)

	resp, err := h.videoClient.GetVideoDetails(r.Context(), &pb.GetVideoDetailsRequest{
		VideoId: videoID,
		UserId:  userID,
	})
	if err != nil {
		http.Error(w, "Video not found", http.StatusNotFound)
		return
	}

	data := map[string]interface{}{
		"Title":         resp.Video.Title + " - TextTube",
		"Authenticated": true,
		"Video":         resp.Video,
	}

	if err := h.templates["video_detail"].ExecuteTemplate(w, "layout.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *SSRHandler) Summarize(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	videoID := vars["videoId"]
	userID := r.Context().Value("user_id").(string)

	resp, err := h.videoClient.SummarizeVideo(r.Context(), &pb.SummarizeVideoRequest{
		VideoId: videoID,
		UserId:  userID,
	})
	if err != nil {
		log.Printf("Summarize error: %v", err)
		http.Redirect(w, r, "/video/"+videoID, http.StatusSeeOther)
		return
	}

	// Fetch video details to render the full page
	videoResp, err := h.videoClient.GetVideoDetails(r.Context(), &pb.GetVideoDetailsRequest{
		VideoId: videoID,
		UserId:  userID,
	})
	if err != nil {
		http.Error(w, "Video not found", http.StatusNotFound)
		return
	}

	data := map[string]interface{}{
		"Title":         videoResp.Video.Title + " - TextTube",
		"Authenticated": true,
		"Video":         videoResp.Video,
		"Summary":       resp.Summary,
	}

	if err := h.templates["video_detail"].ExecuteTemplate(w, "layout.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
