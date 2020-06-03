package parser

import (
	"fmt"
	"github.com/tidwall/gjson"
	"simple-golang-crawler/engine"
	"simple-golang-crawler/fetcher"
	"simple-golang-crawler/model"
	"time"
)

var _getAidUrlTemp = "https://api.bilibili.com/x/space/arc/search?mid=%d&ps=30&tid=0&pn=%d&keyword=&order=pubdate&jsonp=jsonp"
var _getCidUrlTemp = "https://api.bilibili.com/x/player/pagelist?aid=%d"
var createdStart, createdEnd int64

func SetCreatedArea(start string, end string) {
	if start != "" {
		if len(start) == 10 {
			start += " 00:00:00"
		}
		if startTime, err := time.Parse("2006-01-02 15:04:05", start); err == nil {
			createdStart = startTime.Unix()
		} else {
			panic("s参数开始时间错误：" + err.Error())
		}
	}
	if end != "" {
		if len(end) == 10 {
			end += " 00:00:00"
		}
		if endTime, err := time.Parse("2006-01-02 15:04:05", end); err == nil {
			createdEnd = endTime.Unix()
		} else {
			panic("e参数开始时间错误：" + err.Error())
		}
	}
}

func UpSpaceParseFun(contents []byte, url string) engine.ParseResult {
	var retParseResult engine.ParseResult
	value := gjson.GetManyBytes(contents, "data.list.vlist", "data.page")

	var upid int64
	retParseResult.Requests, upid = getAidDetailReqList(value[0])
	retParseResult.Requests = append(retParseResult.Requests, getNewBilibiliUpSpaceReqList(value[1], upid)...)

	return retParseResult

}

func getAidDetailReqList(pageInfo gjson.Result) ([]*engine.Request, int64) {
	var retRequests []*engine.Request
	var upid int64
	for _, i := range pageInfo.Array() {
		upid = i.Get("mid").Int()
		created := i.Get("created").Int()
		if createdStart > 0 && created < createdStart {
			continue
		}
		if createdEnd > 0 && created > createdEnd {
			continue
		}
		aid := i.Get("aid").Int()
		title := i.Get("title").String()
		reqUrl := fmt.Sprintf(_getCidUrlTemp, aid)
		videoAid := model.NewVideoAidInfo(aid, title)
		reqParseFunction := GenGetAidChildrenParseFun(videoAid)
		req := engine.NewRequest(reqUrl, reqParseFunction, fetcher.DefaultFetcher)
		retRequests = append(retRequests, req)
	}
	return retRequests, upid
}

func getNewBilibiliUpSpaceReqList(pageInfo gjson.Result, upid int64) []*engine.Request {
	var retRequests []*engine.Request

	count := pageInfo.Get("count").Int()
	pn := pageInfo.Get("pn").Int()
	ps := pageInfo.Get("ps").Int()
	var extraPage int64
	if count%ps > 0 {
		extraPage = 1
	}
	totalPage := count/ps + extraPage
	for i := int64(1); i < totalPage; i++ {
		if i == pn {
			continue
		}
		reqUrl := fmt.Sprintf(_getAidUrlTemp, upid, i)
		req := engine.NewRequest(reqUrl, UpSpaceParseFun, fetcher.DefaultFetcher)
		retRequests = append(retRequests, req)
	}
	return retRequests
}

func GetRequestByUpId(upid int64) *engine.Request {
	reqUrl := fmt.Sprintf(_getAidUrlTemp, upid, 1)
	return engine.NewRequest(reqUrl, UpSpaceParseFun, fetcher.DefaultFetcher)
}
