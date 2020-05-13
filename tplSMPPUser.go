package main

import "strings"

func tplSMPPUser(userID string) string {
	buf := strings.Builder{}
	buf.WriteString(TplHeaderBegin)
	buf.WriteString(`SMPP User - `)
	buf.WriteString(userID)
	buf.WriteString(` -`)
	buf.WriteString(TplHeaderEnd)
	buf.WriteString(TplNoScript)
	buf.WriteString(`<div class="navbar navbar-fixed-top navbar-default" role="navigation"><div class="container"><div class="navbar-header"><button type="button" class="navbar-toggle" data-toggle="collapse" data-target=".navbar-collapse"><span class="sr-only">Toggle navigation</span><span class="icon-bar"></span><span class="icon-bar"></span><span class="icon-bar"></span></button><a class="navbar-brand" href="./">Beanstalkd console</a></div><div class="collapse navbar-collapse"><ul class="nav navbar-nav">`)
	buf.WriteString(dropDownServer(""))
	buf.WriteString(`</ul><ul class="nav navbar-nav navbar-right"><li class="dropdown"><a href="#" class="dropdown-toggle" data-toggle="dropdown">Toolbox <span class="caret"></span></a><ul class="dropdown-menu"><li><a href="#filter" role="button" data-toggle="modal">Filter columns</a></li><li><a href="./sample?action=manageSamples" role="button">Manage samples</a></li><li><a href="./statistics?action=preference" role="button">Statistics preference</a></li><li class="divider"></li><li><a href="#settings" role="button" data-toggle="modal">Edit settings</a></li></ul></li>`)
	buf.WriteString(TplLinks)
	buf.WriteString(`</div></div></div><div class="container">`)
	buf.WriteString(`<table class='table table-striped table-hover'>`)
	buf.WriteString(findFirstForm(userID, ""))
	buf.WriteString(`</table></div>`)
	return buf.String()
}


// tplFindFirstInTube render navigation search box for search content in jobs by given tube.
func findFirstForm(user, state string) string {
	buf := strings.Builder{}
	buf.WriteString(`<h1>Find SMPP PDUs</h1><p>search query in ready queues</p>`)
	buf.WriteString(`<form role="search" method="post" action="/search">`)
	buf.WriteString(`<input type="hidden" name="state" value="`)
	buf.WriteString(state)
	buf.WriteString(`"/><div class="form-group row"><label for="user_id" class="col-sm-2 col-form-label">User ID</label><div class="col-sm-10"><input type="number" name="user_id" class="form-control" required placeholder="user id" value="`)
	buf.WriteString(user)
	buf.WriteString(`"/></div></div><div class="form-group row"><label for="searchStr" class="col-sm-2 col-form-label">Query</label><div class="col-sm-10"><input type="text" required class="form-control input-sm search-query" name="searchStr" placeholder="query"></div></div><div class="form-group row"><label for="limit" class="col-sm-2 col-form-label">Limit</label><div class="col-sm-10"><input type="number" name="limit" value="25"" placeholder="limit"/></div></div><div class="form-group row"><div class="col-sm-10"><button type="submit" class="btn btn-primary">Search</button></div></div>`)
	buf.WriteString(`</form>`)
	return buf.String()
}


