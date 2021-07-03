package controllers

import (
	"fmt"
	"log"
	"net/http"
	"time"

	. "../utils"
)

func SetPageview(writer http.ResponseWriter, request *http.Request) {
	if request.Method == "POST" {
		request.ParseForm()

		clientid := request.FormValue("clientid")
		guid := request.FormValue("guid")
		path := request.FormValue("path")
		ref := request.FormValue("ref")

		t := time.Now()
		unixTimestamp := t.Unix()

		eventType := "pageview"
		pageviewIndexKey := clientid + KS + "pageviewindex"
		pageViewIndex, err := RC.Do("INCR", pageviewIndexKey)
		if err != nil {
			log.Println(err)
		}
		pageviewHashKey := clientid + KS + eventType + KS + fmt.Sprintf("%v", pageViewIndex)
		pageviewByPathKey := clientid + KS + "path" + KS + path
		pageviewRefKey := clientid + KS + "ref" + KS + ref
		uniquesKey := clientid + KS + "uniques" + KS + t.Format("2006:01:02")
		timeIndexKey := clientid + KS + "timeindex"

		RC.Send("MULTI")
		RC.Send("HMSET", pageviewHashKey, "guid", guid, "path", "ref", ref)
		RC.Send("ZADD", timeIndexKey, unixTimestamp, pageViewIndex)
		RC.Send("ZADD", pageviewByPathKey, unixTimestamp, pageViewIndex)
		RC.Send("ZADD", pageviewRefKey, unixTimestamp, pageViewIndex)
		RC.Send("PFADD", uniquesKey, guid)

		_, err = RC.Do("EXEC")

		if err != nil {
			log.Println(err)
		}
	}
}
