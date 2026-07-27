// /home/krylon/go/src/github.com/blicero/chili/web/web.go
// -*- mode: go; coding: utf-8; -*-
// Created on 22. 07. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-07-26 10:35:37 krylon>

package web

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"log"
	"net/http"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/blicero/chili/common"
	"github.com/blicero/chili/database"
	"github.com/blicero/chili/logdomain"
	"github.com/blicero/chili/model"
	"github.com/gorilla/mux"
)

const (
	// cacheControl = "max-age=3600, public"
	noCache    = "no-store, max-age=0"
	tmplFolder = "assets/templates"
	poolSize   = 4
)

func cacheSeconds(seconds int) string {
	if seconds == 0 {
		return noCache
	}

	return fmt.Sprintf("max-age=%d, public",
		seconds)
} // func cacheSeconds(second int) string

//go:embed assets
var assets embed.FS

// Server wraps the http.Server and associated state.
type Server struct {
	addr      string
	log       *log.Logger
	pool      *database.Pool
	active    atomic.Bool
	lock      sync.RWMutex // nolint: unused
	router    *mux.Router
	tmpl      *template.Template
	web       http.Server
	mimeTypes map[string]string
}

// Create returns a new web Server.
func Create(addr string) (*Server, error) {
	var (
		err error
		msg string
		srv = &Server{
			addr: addr,
			mimeTypes: map[string]string{
				".css":  "text/css",
				".map":  "application/json",
				".js":   "text/javascript",
				".png":  "image/png",
				".jpg":  "image/jpeg",
				".jpeg": "image/jpeg",
				".webp": "image/webp",
				".gif":  "image/gif",
				".json": "application/json",
				".html": "text/html",
			},
		}
	)

	if srv.log, err = common.GetLogger(logdomain.Web); err != nil {
		return nil, err
	} else if srv.pool, err = database.NewPool(poolSize); err != nil {
		srv.log.Printf("[CRITICAL] Failed to open Database connection pool: %s\n",
			err.Error())
		return nil, err
	}

	var templates []fs.DirEntry
	var tmplRe = regexp.MustCompile("[.]tmpl$")

	if templates, err = assets.ReadDir(tmplFolder); err != nil {
		srv.log.Printf("[ERROR] Cannot read embedded templates: %s\n",
			err.Error())
		return nil, err
	}

	srv.tmpl = template.New("").Funcs(funcmap)
	for _, entry := range templates {
		var (
			content []byte
			path    = filepath.Join(tmplFolder, entry.Name())
		)

		if !tmplRe.MatchString(entry.Name()) {
			continue
		} else if content, err = assets.ReadFile(path); err != nil {
			msg = fmt.Sprintf("Cannot read embedded file %s: %s",
				path,
				err.Error())
			srv.log.Printf("[CRITICAL] %s\n", msg)
			return nil, errors.New(msg)
		} else if srv.tmpl, err = srv.tmpl.Parse(string(content)); err != nil {
			msg = fmt.Sprintf("Could not parse template %s: %s",
				entry.Name(),
				err.Error())
			srv.log.Println("[CRITICAL] " + msg)
			return nil, errors.New(msg)
		} else if common.Debug {
			srv.log.Printf("[TRACE] Template \"%s\" was parsed successfully.\n",
				entry.Name())
		}
	}

	srv.router = mux.NewRouter()
	srv.web.Addr = addr
	srv.web.ErrorLog = srv.log
	srv.web.Handler = srv.router

	// Register URL handlers
	srv.router.NotFoundHandler = http.HandlerFunc(srv.handleNotFound)
	srv.router.HandleFunc("/favicon.ico", srv.handleFavIco)
	srv.router.HandleFunc("/static/{file}", srv.handleStaticFile)
	srv.router.HandleFunc("/{index:(?i:index|main|start)$}", srv.handleMain)
	srv.router.HandleFunc("/device/{id:(?:\\d+)$}", srv.handleDeviceDetails)
	// srv.router.HandleFunc("/host/all", srv.handleHostsView)
	// srv.router.HandleFunc("/host/{name}/chart", srv.handleHostChart)
	// srv.router.HandleFunc("/host/{name}", srv.handleSingleHostView)

	// AJAX Handlers
	srv.router.HandleFunc(
		"/ajax/beacon",
		srv.handleBeacon)

	// Web service endpoints
	// srv.router.HandleFunc("/ws/get_timestamp/{name:(?:\\w+)$}",
	// 	srv.handleClientGetTimestamp)

	// srv.router.HandleFunc("/ws/submit_data/{name:(?:\\w+)$}",
	// 	srv.handleClientData)

	return srv, nil
} // func Create(addr string) (*Server, error)

// IsActive returns the Server's active flag.
func (srv *Server) IsActive() bool {
	return srv.active.Load()
} // func (srv *Server) IsActive() bool

// Stop clears the Server's active flag.
func (srv *Server) Stop() {
	srv.active.Store(false)
	srv.web.Shutdown(context.Background()) // nolint: errcheck
	srv.pool.Close()                       // nolint: errcheck
} // func (srv *Server) Stop()

// Run executes the Server's loop, waiting for new connections and starting
// goroutines to handle them.
func (srv *Server) Run() {
	var err error

	defer srv.log.Println("[INFO] Web server is shutting down")

	srv.active.Store(true)
	defer srv.active.Store(false)

	srv.log.Printf("[INFO] Web frontend is going online at %s\n", srv.addr)
	http.Handle("/", srv.router)

	if err = srv.web.ListenAndServe(); err != nil {
		if err.Error() != "http: Server closed" {
			srv.log.Printf("[ERROR] ListenAndServe returned an error: %s\n",
				err.Error())
		} else {
			srv.log.Println("[INFO] HTTP Server has shut down.")
		}
	}
} // func (srv *Server) Run()

//////////////////////////////////////////////////////////////////////////////
/// Handle requests //////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////

func (srv *Server) handleMain(w http.ResponseWriter, r *http.Request) {
	srv.log.Printf("[TRACE] Handling request for %s\n", r.RequestURI)
	const tmplName = "main"

	var (
		err  error
		msg  string
		db   *database.Database
		tmpl *template.Template
		data = tmplDataIndex{
			tmplDataBase: tmplDataBase{
				Title: "Main",
				Debug: common.Debug,
				URL:   r.URL.String(),
			},
		}
	)

	db = srv.pool.Get()
	defer srv.pool.Put(db)

	if tmpl = srv.tmpl.Lookup(tmplName); tmpl == nil {
		msg = fmt.Sprintf("Could not find template %q", tmplName)
		srv.log.Println("[CRITICAL] " + msg)
		srv.sendErrorMessage(w, msg)
		return
	} else if data.Networks, err = db.NetGetAll(); err != nil {
		msg = fmt.Sprintf("failed to load Networks: %s", err.Error())
		srv.log.Println("[CRITICAL] " + msg)
		srv.sendErrorMessage(w, msg)
		return
	} else if data.Devices, err = db.DeviceGetAll(); err != nil {
		msg = fmt.Sprintf("failed to load Devices: %s", err.Error())
		srv.log.Println("[CRITICAL] " + msg)
		srv.sendErrorMessage(w, msg)
		return
	}

	slices.SortFunc(data.Devices, func(a, b *model.Device) int {
		if a.Name < b.Name {
			return -1
		} else if a.Name > b.Name {
			return 1
		}
		return 0
	})

	w.Header().Set("Cache-Control", noCache)
	if err = tmpl.Execute(w, &data); err != nil {
		msg = fmt.Sprintf("Error rendering template %q: %s",
			tmplName,
			err.Error())
		srv.sendErrorMessage(w, msg)
	}
} // func (srv *Server) handleMain(w http.ResponseWriter, r *http.Request)

func (srv *Server) handleDeviceDetails(w http.ResponseWriter, r *http.Request) {
	const tmplName = "device_details"
	var (
		err        error
		msg, idStr string
		devID      int64
		vars       map[string]string
		db         *database.Database
		tmpl       *template.Template
		data       = tmplDataDeviceDetails{
			tmplDataBase: tmplDataBase{
				Debug: common.Debug,
				URL:   r.URL.String(),
			},
		}
	)

	vars = mux.Vars(r)
	idStr = vars["id"]

	if devID, err = strconv.ParseInt(idStr, 10, 64); err != nil {
		msg = fmt.Sprintf("Could not parse ID %q", idStr)
		srv.log.Println("[ERROR] " + msg)
		srv.sendErrorMessage(w, msg)
		return
	}

	db = srv.pool.Get()
	defer srv.pool.Put(db)

	if data.Device, err = db.DeviceGetByID(devID); err != nil {
		msg = fmt.Sprintf("Error looking up device %d: %s",
			devID,
			err.Error())
		srv.log.Printf("[ERROR] %s\n", msg)
		srv.sendErrorMessage(w, msg)
		return
	} else if data.Device == nil {
		msg = fmt.Sprintf("Did not find Device %d in database",
			devID)
		srv.log.Printf("[ERROR] %s\n", msg)
		srv.sendErrorMessage(w, msg)
		return
	} else if data.Attributes, err = db.AttributeGetMostRecent(data.Device); err != nil {
		msg = fmt.Sprintf("Error looking for Attributes of %s: %s",
			data.Device.Name,
			err.Error())
		srv.log.Printf("[ERROR] %s\n", msg)
		srv.sendErrorMessage(w, msg)
		return
	} else if tmpl = srv.tmpl.Lookup(tmplName); tmpl == nil {
		msg = fmt.Sprintf("Could not find template %q", tmplName)
		srv.log.Println("[CRITICAL] " + msg)
		srv.sendErrorMessage(w, msg)
		return
	}

	data.Title = fmt.Sprintf("Device Details for %s",
		data.Device.Name)

	w.Header().Set("Cache-Control", noCache)
	if err = tmpl.Execute(w, &data); err != nil {
		msg = fmt.Sprintf("Error rendering template %q: %s",
			tmplName,
			err.Error())
		srv.sendErrorMessage(w, msg)
	}
} // func (srv *Server) handleDeviceDetails(w http.ResponseWriter, r *http.Request)

func (srv *Server) handleNotFound(w http.ResponseWriter, r *http.Request) {
	srv.log.Printf("[TRACE] Handling request for %s\n", r.RequestURI)
	srv.log.Printf("[ERROR] 404 - %s\n", r.RequestURI)

	srv.sendErrorMessage(
		w,
		fmt.Sprintf(
			"No Handler was found for %s",
			r.RequestURI))
} // func (srv *Server) handleNotFound(w http.ResponseWriter, r *http.Request)

//////////////////////////////////////////////////////////////////////////////
/// Handle AJAX //////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////

func (srv *Server) handleBeacon(w http.ResponseWriter, r *http.Request) {
	var (
		err  error
		buf  []byte
		data = ajaxBeaconData{
			ajaxData: ajaxData{
				Status:    true,
				Timestamp: time.Now(),
				Message:   common.AppName + " " + common.Version,
			},
			Hostname: hostname(),
		}
	)

	if buf, err = json.Marshal(&data); err != nil {
		var msg = fmt.Sprintf("Failed to serialize payload for AJAX response: %s",
			err.Error())
		srv.log.Printf("[CANTHAPPEN] %s\n", msg)
		buf = errJSON(msg)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", noCache)
	w.WriteHeader(200)
	w.Write(buf) // nolint: errcheck,gosec
} // func (srv *Server) handleBeacon(w http.ResponseWriter, r *http.Request)

//////////////////////////////////////////////////////////////////////////////
/// Handle static assets /////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////

func (srv *Server) handleFavIco(w http.ResponseWriter, request *http.Request) {
	srv.log.Printf("[TRACE] Handle request for %s\n",
		request.URL.EscapedPath())

	const (
		filename = "assets/static/favicon.ico"
		mimeType = "image/vnd.microsoft.icon"
	)

	w.Header().Set("Content-Type", mimeType)

	// if !common.Debug {
	// 	w.Header().Set("Cache-Control", cacheControl)
	// } else {
	// 	w.Header().Set("Cache-Control", noCache)
	// }
	w.Header().Set("Cache-Control", cacheSeconds(900))

	var (
		err error
		fh  fs.File
	)

	if fh, err = assets.Open(filename); err != nil {
		msg := fmt.Sprintf("ERROR - cannot find file %s", filename)
		srv.sendErrorMessage(w, msg)
	} else {
		defer fh.Close() // nolint: errcheck
		w.WriteHeader(200)
		io.Copy(w, fh) // nolint: errcheck
	}
} // func (srv *Server) handleFavIco(w http.ResponseWriter, request *http.Request)

func (srv *Server) handleStaticFile(w http.ResponseWriter, request *http.Request) {
	// srv.log.Printf("[TRACE] Handle request for %s\n",
	// 	request.URL.EscapedPath())

	// Since we controll what static files the server has available, we
	// can easily map MIME type to slice. Soon.

	vars := mux.Vars(request)
	filename := vars["file"]
	path := filepath.Join("assets", "static", filename)

	var mimeType string

	var match []string

	if match = common.SuffixPattern.FindStringSubmatch(filename); match == nil {
		mimeType = "text/plain"
	} else if mime, ok := srv.mimeTypes[match[1]]; ok {
		mimeType = mime
	} else {
		srv.log.Printf("[ERROR] Did not find MIME type for %s\n", filename)
	}

	w.Header().Set("Content-Type", mimeType)
	w.Header().Set("Cache-Control", cacheSeconds(900))

	// if common.Debug {
	// 	w.Header().Set("Cache-Control", noCache)
	// } else {
	// 	w.Header().Set("Cache-Control", cacheControl)
	// }

	var (
		err error
		fh  fs.File
	)

	if fh, err = assets.Open(path); err != nil {
		msg := fmt.Sprintf("ERROR - cannot find file %s", path)
		srv.sendErrorMessage(w, msg)
	} else {
		defer fh.Close() // nolint: errcheck
		w.WriteHeader(200)
		io.Copy(w, fh) // nolint: errcheck
	}
} // func (srv *Server) handleStaticFile(w http.ResponseWriter, request *http.Request)

func (srv *Server) sendErrorMessage(w http.ResponseWriter, msg string) {
	html := `
<!DOCTYPE html>
<html>
  <head>
    <title>Internal Error</title>
  </head>
  <body>
    <h1>Internal Error</h1>
    <hr />
    We are sorry to inform you an internal application error has occured:<br />
    %s
    <p>
    Back to <a href="/index">Homepage</a>
    <hr />
    &copy; 2026 <a href="mailto:krylon@gmx.net">Benjamin Walkenhorst</a>
  </body>
</html>
`

	w.Header().Set("Cache-Control", noCache)
	srv.log.Printf("[ERROR] %s\n", msg)

	output := fmt.Sprintf(html, msg)
	w.WriteHeader(500)
	_, _ = w.Write([]byte(output)) // nolint: gosec
} // func (srv *Server) sendErrorMessage(w http.ResponseWriter, msg string)
