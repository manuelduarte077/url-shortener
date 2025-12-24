package handlers

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"strings"
	"url-shortener/internal/application"
	"url-shortener/internal/infrastructure/qrcode"
)

type ShortenerHandler struct {
	service     *application.ShortenerService
	tmpl        *template.Template
	qrGenerator *qrcode.QRCodeGenerator
}

func NewShortenerHandler(service *application.ShortenerService, tmpl *template.Template) *ShortenerHandler {
	return &ShortenerHandler{
		service:     service,
		tmpl:        tmpl,
		qrGenerator: qrcode.NewQRCodeGenerator(),
	}
}

func (h *ShortenerHandler) ShowForm(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := h.tmpl.ExecuteTemplate(w, "form.html", nil); err != nil {
		log.Printf("Error rendering form template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (h *ShortenerHandler) CreateShortURL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	longURL := r.FormValue("url")
	if longURL == "" {
		http.Error(w, "URL is required", http.StatusBadRequest)
		return
	}

	if _, err := url.ParseRequestURI(longURL); err != nil {
		http.Error(w, "Invalid URL format", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	shortURL, err := h.service.CreateShortURL(ctx, longURL)
	if err != nil {
		log.Printf("Error creating short URL: %v", err)
		http.Error(w, "Failed to create short URL", http.StatusInternalServerError)
		return
	}

	data := struct {
		ShortCode string
		LongURL   string
		ShortURL  string
	}{
		ShortCode: shortURL.ShortCode,
		LongURL:   shortURL.LongURL,
		ShortURL:  h.buildShortURL(r, shortURL.ShortCode),
	}

	if err := h.tmpl.ExecuteTemplate(w, "result.html", data); err != nil {
		log.Printf("Error rendering result template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (h *ShortenerHandler) Redirect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	shortCode := r.URL.Path[1:]
	if shortCode == "" {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	ctx := r.Context()
	longURL, err := h.service.GetLongURL(ctx, shortCode)
	if err != nil {
		log.Printf("Error getting long URL: %v", err)
		http.Error(w, "URL not found", http.StatusNotFound)
		return
	}

	http.Redirect(w, r, longURL, http.StatusMovedPermanently)
}

func (h *ShortenerHandler) GetQRCode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/qrcode/")
	if path == "" {
		http.Error(w, "Short code is required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	_, err := h.service.GetLongURL(ctx, path)
	if err != nil {
		log.Printf("Error getting long URL for QR code: %v", err)
		http.Error(w, "URL not found", http.StatusNotFound)
		return
	}

	shortURL := h.buildShortURL(r, path)

	pngData, err := h.qrGenerator.GeneratePNG(shortURL)
	if err != nil {
		log.Printf("Error generating QR code: %v", err)
		http.Error(w, "Failed to generate QR code", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(pngData)))
	w.Header().Set("Cache-Control", "public, max-age=3600")

	if _, err := w.Write(pngData); err != nil {
		log.Printf("Error writing QR code response: %v", err)
	}
}

func (h *ShortenerHandler) buildShortURL(r *http.Request, shortCode string) string {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s/%s", scheme, r.Host, shortCode)
}
