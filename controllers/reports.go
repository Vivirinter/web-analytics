package controllers

import (
	. "../utils"
	"github.com/gomodule/redigo/redis"
	"github.com/unrolled/render"
	"log"
	"net/http"
	"strings"
)

type pageviews struct {
	Views     int64
	StartTime string
	EndTime   string
	Path      string
	Ref       string
}

type uniques struct {
	Daily   int64
	Monthly int64
	Yearly  int64
	Time    string
}

var clientid = "287954214567" //This would be fetched from the DB or the client's session
var pv *pageviews
var r = render.New()

func GetUniques(writer http.ResponseWriter, request *http.Request) {

	if request.Method == "GET" {
		d := request.URL.Query().Get("d")

		dateComponents := strings.Split(d, "/")
		if len(dateComponents) <= 1 {
			r.Text(writer, http.StatusOK, "Error: Please pass in a correctly formatted date: YYYY/MM/DD")
		} else {
			dailyUniques, err := RC.Do("PFCOUNT", clientid+KS+"uniques"+KS+dateComponents[0]+KS+dateComponents[1]+KS+dateComponents[2]) //PFCOUNT on the hyperloglog object for daily uniques
			if err != nil {
				log.Println(err)
			}
			monthlyUniques := queryUniques(dateComponents[0]+KS+dateComponents[1], "monthlyuniques")
			yearlyUniques := queryUniques(dateComponents[0], "yearlyuniques")

			uniques := uniques{
				Daily:   dailyUniques.(int64),
				Monthly: monthlyUniques.(int64),
				Yearly:  yearlyUniques.(int64),
				Time:    d,
			}
			r.JSON(writer, http.StatusOK, uniques)
		}

	}
}

func queryUniques(keypattern string, resultkey string) interface{} {

	//Clear out any old results
	_, err := RC.Do("DEL", clientid+KS+resultkey)
	if err != nil {
		log.Println(err)
	}

	matchingKeys := getMatchingKeys(clientid + KS + "uniques" + KS + keypattern + KS + "*")

	combniedKeys := append([]string{clientid + KS + resultkey}, matchingKeys...)

	s := make([]interface{}, len(combniedKeys))
	for index, value := range combniedKeys {
		s[index] = value
	}

	_, err = RC.Do("PFMERGE", s...)
	if err != nil {
		log.Println(err)
	}
	uniques, err := RC.Do("PFCOUNT", clientid+KS+resultkey)
	if err != nil {
		log.Println(err)
	}

	return uniques
}

func getMatchingKeys(pattern string) []string {
	iter := 0

	keys := []string{}
	for {

		arr, err := redis.Values(RC.Do("SCAN", 0, "MATCH", pattern, "COUNT", 365))
		if err != nil {
			log.Println(err)
		} else {
			iter, err = redis.Int(arr[0], nil)
			if err != nil {
				log.Println(err)
			}

			k, err := redis.Strings(arr[1], nil)
			if err != nil {
				log.Println(err)
			}

			keys = append(keys, k...)
		}

		if iter == 0 {
			break
		}
	}
	return keys
}

func GetPageviews(writer http.ResponseWriter, request *http.Request) {

	if request.Method == "GET" {

		getParams := pageviews{
			Views:     0,
			StartTime: request.URL.Query().Get("s"),
			EndTime:   request.URL.Query().Get("e"),
			Path:      request.URL.Query().Get("path"),
			Ref:       request.URL.Query().Get("ref"),
		}

		if getParams.Path != "" || getParams.Ref != "" {

			if getParams.Path != "" && getParams.Ref != "" {
				pv = getPageviewsByRefAndPath(&getParams)
			}

			if getParams.Path == "" && getParams.Ref != "" {
				pv = getPageviewsByRef(&getParams)
			}

			if getParams.Path != "" && getParams.Ref == "" {
				pv = getPageviewsByPath(&getParams)
			}

		} else {
			pv = getAllPageViews(&getParams)
		}

		r.JSON(writer, http.StatusOK, pv)
	}

}

func getPageviewsByRefAndPath(pv *pageviews) *pageviews {
	hits, err := RC.Do("ZINTERSTORE", "out", 2, clientid+KS+"path"+KS+pv.Path, clientid+KS+"ref"+KS+pv.Ref, "WEIGHTS", pv.StartTime, pv.EndTime)
	if err != nil {
		log.Println(err)
	}
	pv.Views = hits.(int64)
	return pv
}

func getPageviewsByRef(pv *pageviews) *pageviews {
	hits, err := RC.Do("ZCOUNT", clientid+KS+"ref"+KS+pv.Ref, pv.StartTime, pv.EndTime)
	if err != nil {
		log.Println(err)
	}
	pv.Views = hits.(int64)
	pv.Path = "ALL"
	return pv
}

func getPageviewsByPath(pv *pageviews) *pageviews {
	hits, err := RC.Do("ZCOUNT", clientid+KS+"path"+KS+pv.Path, pv.StartTime, pv.EndTime)
	if err != nil {
		log.Println(err)
	}
	pv.Views = hits.(int64)
	pv.Ref = "ALL"
	return pv
}

func getAllPageViews(pv *pageviews) *pageviews {
	hits, err := RC.Do("ZCOUNT", clientid+KS+"timeindex", pv.StartTime, pv.EndTime)
	if err != nil {
		log.Println(err)
	}
	pv.Views = hits.(int64)
	pv.Path = "ALL"
	pv.Ref = "ALL"
	return pv
}
