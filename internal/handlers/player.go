package handlers

import (
	"html/template"
	"net/http"
	"strconv"

	"music-session-app/internal/auth"
	"music-session-app/internal/middleware"

	"github.com/alexedwards/scs/v2"
)

type PlayerHandler struct {
	Sessions *scs.SessionManager
}

// catalog is the fixed set of tracks the player can play. Using a server-side
// catalog (rather than listing the static dir) keeps display names tidy and
// avoids exposing arbitrary files under /static.
type Track struct {
	File   string
	Title  string
	Artist string
}

var catalog = []Track{
	{File: "track1.wav", Title: "Sunrise Arpeggio", Artist: "Resonate Demos"},
	{File: "track2.wav", Title: "Midnight Pulse", Artist: "Resonate Demos"},
	{File: "track3.wav", Title: "Coast Drive", Artist: "Resonate Demos"},
}

// posKey is the session key under which we store the resume position (seconds)
// for a given track file.
func posKey(file string) string { return "pos:" + file }

// isKnownTrack guards against saving progress for files outside the catalog.
func isKnownTrack(file string) bool {
	for _, t := range catalog {
		if t.File == file {
			return true
		}
	}
	return false
}

type TrackView struct {
	File     string
	Title    string
	Artist   string
	Position int  // saved resume position in seconds
	Current  bool // whether this is the track to load first
}

type PlayerData struct {
	Username string
	Tracks   []TrackView
}

func (h *PlayerHandler) ShowPlayer(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value(middleware.ClaimsKey).(*auth.Claims)

	// Which track was the user last on? Default to the first in the catalog.
	lastSong := h.Sessions.GetString(r.Context(), "last_song")
	if !isKnownTrack(lastSong) {
		lastSong = catalog[0].File
	}

	tracks := make([]TrackView, 0, len(catalog))
	for _, t := range catalog {
		tracks = append(tracks, TrackView{
			File:     t.File,
			Title:    t.Title,
			Artist:   t.Artist,
			Position: h.Sessions.GetInt(r.Context(), posKey(t.File)),
			Current:  t.File == lastSong,
		})
	}

	tmpl := template.Must(template.ParseFiles("templates/player.html"))
	tmpl.Execute(w, PlayerData{
		Username: claims.Username,
		Tracks:   tracks,
	})
}

func (h *PlayerHandler) SaveProgress(w http.ResponseWriter, r *http.Request) {
	song := r.FormValue("song")
	if !isKnownTrack(song) {
		http.Error(w, "unknown track", http.StatusBadRequest)
		return
	}

	position, err := strconv.Atoi(r.FormValue("position"))
	if err != nil || position < 0 {
		position = 0
	}

	// Remember which track we were on, and the per-track resume position.
	h.Sessions.Put(r.Context(), "last_song", song)
	h.Sessions.Put(r.Context(), posKey(song), position)

	w.WriteHeader(http.StatusOK)
}
