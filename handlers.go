// Copyright 2016 - 2020 The aurora Authors. All rights reserved. Use of this
// source code is governed by a MIT license that can be found in the LICENSE
// file.
//
// The aurora is a web-based beanstalkd queue server console written in Go
// and works on macOS, Linux and Windows machines. Main idea behind using Go
// for backend development is to utilize ability of the compiler to produce
// zero-dependency binaries for multiple platforms. aurora was created as an
// attempt to build very simple and portable application to work with local or
// remote beanstalkd server.

package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

// handlerMain handle request on router: /
func handlerMain(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Server", "Go WebServer")
	w.Header().Set("Content-Type", "text/html")
	server := r.URL.Query().Get("server")
	readCookies(r)
	_, _ = io.WriteString(w, tplMain(getServerStatus(), server))
}

// handlerServerList handle request on router: /index
func handlerServerList(w http.ResponseWriter, r *http.Request) {
	setHeader(w, r)
	readCookies(r)
	_, _ = io.WriteString(w, getServerStatus())
}

// serversRemove handle request on router: /serversRemove
func serversRemove(w http.ResponseWriter, r *http.Request) {
	setHeader(w, r)
	readCookies(r)
	server := r.URL.Query().Get("removeServer")
	removeServerInCookie(server, w, r)
	removeServerInConfig(server)
	w.Header().Set("Location", "./public")
	w.WriteHeader(307)
}

// handlerServer handle request on router: /server
func handlerServer(w http.ResponseWriter, r *http.Request) {
	setHeader(w, r)
	readCookies(r)
	server := r.URL.Query().Get("server")
	action := r.URL.Query().Get("action")
	switch action {
	case "reloader":
		_, _ = io.WriteString(w, getServerTubes(server))
		return
	case "clearTubes":
		_ = r.ParseForm()
		clearTubes(server, r.Form)
		_, _ = io.WriteString(w, `{"result":true}`)
		return
	}
	_, _ = io.WriteString(w, tplServer(getServerTubes(server), server))
}

// handlerTube handle request on router: /tube
func handlerTube(w http.ResponseWriter, r *http.Request) {
	setHeader(w, r)
	readCookies(r)
	server := r.URL.Query().Get("server")
	tube := r.URL.Query().Get("tube")
	action := r.URL.Query().Get("action")
	count := r.URL.Query().Get("count")
	switch action {
	case "addjob":
		addJob(server, r.PostFormValue("tubeName"), r.PostFormValue("tubeData"), r.PostFormValue("tubePriority"), r.PostFormValue("tubeDelay"), r.PostFormValue("tubeTtr"))
		_, _ = io.WriteString(w, `{"result":true}`)
		return
	case "search":
		content := searchTube(server, tube, r.URL.Query().Get("limit"), r.URL.Query().Get("searchStr"))
		_, _ = io.WriteString(w, tplTube(content, server, tube))
		return
	case "addSample":
		_ = r.ParseForm()
		addSample(server, r.Form, w)
		return
	default:
		handleRedirect(w, r, server, tube, action, count)
	}
}

// handleRedirect handle request with redirect response.
func handleRedirect(w http.ResponseWriter, r *http.Request, server string, tube string, action string, count string) {
	var link strings.Builder
	link.WriteString(`./tube?server=`)
	link.WriteString(server)
	link.WriteString(`&tube=`)
	switch action {
	case "kick":
		kick(server, tube, count)
		link.WriteString(url.QueryEscape(tube))
		w.Header().Set("Location", link.String())
		w.WriteHeader(307)
	case "kickJob":
		kickJob(server, tube, r.URL.Query().Get("jobid"))
		link.WriteString(url.QueryEscape(tube))
		w.Header().Set("Location", link.String())
		w.WriteHeader(307)
	case "pause":
		pause(server, tube, count)
		link.WriteString(url.QueryEscape(tube))
		w.Header().Set("Location", link.String())
		w.WriteHeader(307)
	case "moveJobsTo":
		destTube := tube
		if r.URL.Query().Get("destTube") != "" {
			destTube = r.URL.Query().Get("destTube")
		}
		moveJobsTo(server, tube, destTube, r.URL.Query().Get("state"), r.URL.Query().Get("destState"))
		link.WriteString(url.QueryEscape(destTube))
		w.Header().Set("Location", link.String())
		w.WriteHeader(307)
	case "deleteAll":
		deleteAll(server, tube)
		link.WriteString(url.QueryEscape(tube))
		w.Header().Set("Location", link.String())
		w.WriteHeader(307)
	case "deleteJob":
		deleteJob(server, tube, r.URL.Query().Get("jobid"))
		link.WriteString(url.QueryEscape(tube))
		w.Header().Set("Location", link.String())
		w.WriteHeader(307)
	case "loadSample":
		loadSample(server, tube, r.URL.Query().Get("key"))
		link.WriteString(url.QueryEscape(tube))
		w.Header().Set("Location", link.String())
		w.WriteHeader(307)
	}
	_, _ = io.WriteString(w, tplTube(currentTube(server, tube), server, tube))
}

// handlerSample handle request on router: /sample
func handlerSample(w http.ResponseWriter, r *http.Request) {
	setHeader(w, r)
	readCookies(r)
	action := r.URL.Query().Get("action")
	server := r.URL.Query().Get("server")
	switch action {
	case "manageSamples":
		_, _ = io.WriteString(w, tplSampleJobsManage(getSampleJobList(), server))
		return
	case "newSample":
		_, _ = io.WriteString(w, tplSampleJobsManage(tplSampleJobEdit("", ""), server))
		return
	case "editSample":
		_, _ = io.WriteString(w, tplSampleJobsManage(tplSampleJobEdit(r.URL.Query().Get("key"), ""), server))
		return
	case "actionNewSample":
		_ = r.ParseForm()
		newSample(server, r.Form, w, r)
		return
	case "actionEditSample":
		_ = r.ParseForm()
		editSample(server, r.Form, r.URL.Query().Get("key"), w, r)
		return
	case "deleteSample":
		deleteSamples(r.URL.Query().Get("key"))
		w.Header().Set("Location", "./sample?action=manageSamples")
		w.WriteHeader(307)
		return
	}
}

// handlerStatistics handle request on router: /statistics
func handlerStatistics(w http.ResponseWriter, r *http.Request) {
	setHeader(w, r)
	readCookies(r)
	action := r.URL.Query().Get("action")
	server := r.URL.Query().Get("server")
	tube := r.URL.Query().Get("tube")
	switch action {
	case "preference":
		_, _ = io.WriteString(w, tplStatisticSetting(tplStatisticEdit("")))
		return
	case "save":
		_ = r.ParseForm()
		statisticPreferenceSave(r.Form, w, r)
		return
	case "reloader":
		_, _ = io.WriteString(w, statisticWaitress(server, tube))
		return
	}
	_, _ = io.WriteString(w, tplStatistic(server, tube))
}

func handlerSMPPUser(w http.ResponseWriter, r *http.Request) {
	setHeader(w, r)
	readCookies(r)

	userID := r.URL.Query().Get("user_id")

	_, _ = io.WriteString(w, tplSMPPUser(userID))

}

func handlerSMPPSearch(w http.ResponseWriter, r *http.Request) {
	setHeader(w, r)
	readCookies(r)

	switch r.Method {
	case http.MethodPost:
		var jobID string
		var err error
		var searchLimit int
		searchLimit, err = strconv.Atoi(r.PostFormValue("limit"))
		if err != nil {
			_, _ = io.WriteString(w, fmt.Sprintf("error: %s", err))
			return
		}
		userID := r.PostFormValue("user_id")

		// check if queue exists
		if !UserTubeExist(userID) {
			_, _ = io.WriteString(w, fmt.Sprintf("tube for user %s does not exist", userID))
			return
		}

		if jobID, err = EnqueueSearchJob(r.PostFormValue("user_id"), r.PostFormValue("state"), r.PostFormValue("searchStr"), searchLimit); err != nil {
			_, _ = io.WriteString(w, fmt.Sprintf("error: %s", err))
			return
		}
		http.Redirect(w, r, fmt.Sprintf("/search?job_id=%s", jobID), http.StatusSeeOther)
		return
	case http.MethodGet:
		buf := strings.Builder{}
		buf.WriteString(TplHeaderBegin)
		buf.WriteString(TplHeaderEnd)
		buf.WriteString(TplNoScript)
		buf.WriteString(`<div class="navbar navbar-fixed-top navbar-default" role="navigation"><div class="container"><div class="navbar-header"><button type="button" class="navbar-toggle" data-toggle="collapse" data-target=".navbar-collapse"><span class="sr-only">Toggle navigation</span><span class="icon-bar"></span><span class="icon-bar"></span><span class="icon-bar"></span></button><a class="navbar-brand" href="./">Beanstalkd console</a></div><div class="collapse navbar-collapse"><ul class="nav navbar-nav">`)
		buf.WriteString(dropDownServer(""))
		buf.WriteString(`</ul><ul class="nav navbar-nav navbar-right"><li class="dropdown"><a href="#" class="dropdown-toggle" data-toggle="dropdown">Toolbox <span class="caret"></span></a><ul class="dropdown-menu"><li><a href="#filter" role="button" data-toggle="modal">Filter columns</a></li><li><a href="./sample?action=manageSamples" role="button">Manage samples</a></li><li><a href="./statistics?action=preference" role="button">Statistics preference</a></li><li class="divider"></li><li><a href="#settings" role="button" data-toggle="modal">Edit settings</a></li></ul></li>`)
		buf.WriteString(TplLinks)
		buf.WriteString(`</div></div></div><div class="container">`)

		jobID := r.URL.Query().Get("job_id")
		if jobID == "" {
			buf.WriteString(`<table class='table table-striped table-hover'>`)
			for jobID, results := range searchResults {
				buf.WriteString(fmt.Sprintf("<tr><td><a href='/search?job_id=%s'>job #ID: %s</a></td><td>%s</td></tr>", jobID, jobID, results.Status))
			}
			buf.WriteString(`</table></div>`)
		} else {
			if results, ok := searchResults[jobID]; ok {
				buf.WriteString(`<table class='table table-striped table-hover'>`)
				buf.WriteString(fmt.Sprintf("<tr><td>ID#%s</td><td>Status: %s</td><td>Created at: %s</td></tr>", results.ID, results.Status, results.CreatedAt.Format(time.UnixDate)))
				buf.WriteString(`</table>`)
				buf.WriteString(tplSearchResultView(currentTubeSearchResults("", results.Tube, "1", results.Query, results.Result), "", results.Tube))
			} else {
				buf.WriteString(fmt.Sprintf("<p>jobID %s not found</p>", jobID))
			}
		}

		_, _ = io.WriteString(w, buf.String())
	}
}


func tplSearchResultView(content string, server string, tube string) string {
	var isDisabledJobDataHighlight string
	if selfConf.IsDisabledJobDataHighlight != 1 {
		isDisabledJobDataHighlight = `<script src="./highlight/highlight.pack.js"></script><script>hljs.initHighlightingOnLoad();</script>`
	}
	buf := strings.Builder{}
	buf.WriteString(content)
	buf.WriteString(modalAddJob(tube))
	buf.WriteString(modalAddSample(server, tube))
	buf.WriteString(`<div id="idAllTubesCopy" style="display:none"></div>`)
	buf.WriteString(tplTubeFilter())
	buf.WriteString(dropEditSettings())
	buf.WriteString(`</div><script>function getParameterByName(name,url){if(!url){url=window.location.href}name=name.replace(/[\[\]]/g,"\\$&");var regex=new RegExp("[?&]"+name+"(=([^&#]*)|&|#|$)"),results=regex.exec(url);if(!results){return null}if(!results[2]){return""}return decodeURIComponent(results[2].replace(/\+/g," "))}var url="./tube?server="+getParameterByName("server");var contentType="";</script><script src='./assets/vendor/jquery/jquery.js'></script><script src="./js/jquery.color.js"></script><script src="./js/jquery.cookie.js"></script><script src="./js/jquery.regexp.js"></script><script src="./assets/vendor/bootstrap/js/bootstrap.min.js"></script>`)
	buf.WriteString(isDisabledJobDataHighlight)
	buf.WriteString(`<script src="./js/customer.js"></script></body></html>`)
	return buf.String()
}


var jobs chan *SearchJob
var searchResults map[string]*SearchResults

type SearchJob struct {
	Tube  string
	Query string
	State string
	ID    string
	Limit int
}

func init() {
	// jobs
	jobs = make(chan *SearchJob)
	searchResults = make(map[string]*SearchResults)
}

func UserTube(userID string) string {
	return fmt.Sprintf("mt-sms-smpp-out-%s", userID)
}

func UserTubeExist(userID string) bool {
	 for _, server := range selfConf.Servers {
		 if exists, err := TubeExist(server, UserTube(userID)); err != nil {
		 	log.Printf("failed to check tube existing: %s", err)
		 	return false
		 } else if exists {
		 	return true
		 }
	 }

	 return false
}

func NewSearchInUserTubes(userID, state, searchStr string, searchLimit int) *SearchJob {
	id, _ := uuid.NewUUID()
	return &SearchJob{ID: id.String(), Tube: UserTube(userID), State: state, Query: searchStr, Limit: searchLimit}
}

func EnqueueSearchJob(userID, state, searchStr string, searchLimit int) (string, error) {
	job := NewSearchInUserTubes(userID, state, searchStr, searchLimit)

	jobs <- job

	return job.ID, nil
}


type SearchResults struct {
	Result    []SearchResult
	ID        string
	Status    string
	Tube      string
	Query     string
	CreatedAt time.Time
	Limit     int
}

func ProcessSearchJob(ctx context.Context) error {
	for {
		select {
		case job := <-jobs:
			searchResults[job.ID] = &SearchResults{Result: []SearchResult{}, ID: job.ID, Status: "pending", Tube: job.Tube, Query: job.Query, CreatedAt: time.Now(), Limit: job.Limit}

			var results []SearchResult
			for _, server := range selfConf.Servers {
				if res, err := searchTubeWithReadyState(server, job.Tube, job.Limit, job.Query); err != nil {
					log.Printf("error during job processment: %s", err)
					continue
				} else {
					results = append(results, res...)
				}
			}
			searchResults[job.ID].Result = results
			searchResults[job.ID].Status = "finished"
		case <-ctx.Done():
			return nil
		}
	}

	return nil
}