// Copyright 2011 Dmitry Chestnykh. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// example of HTTP server that uses the captcha package.
package main

import (
	"encoding/json"
	"fmt"
	"github.com/mathisonqin/captcha"
	"github.com/miguel-branco/goconfig"
	"io"
	"log"
	_ "net/http/pprof"
	"net/http"
	"runtime"
	"strconv"
	"text/template"
)

const (
	HomeRoot = "/home/captcha/captcha"
)

var formTemplate = template.Must(template.New("example").Parse(formTemplateSrc))

func showFormHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/captcha/" {
		http.NotFound(w, r)
		return
	}
	d := struct {
		CaptchaId string
	}{
		captcha.New(),
	}
	if err := formTemplate.Execute(w, &d); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func processFormHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if !captcha.VerifyString(r.FormValue("captchaId"), r.FormValue("captchaValue")) {
		io.WriteString(w, "Wrong captcha solution! No robots allowed!\n")
	} else {
		io.WriteString(w, "Great job, human! You solved the captcha.\n")
	}
	io.WriteString(w, "<br><a href='/captcha'>Try another one</a>")
}

func exception(w http.ResponseWriter) {
	if err := recover(); err != nil {
		res := struct {
			Code int
			Msg  string
		}{
			1,
			"internal error",
		}
		strRes, _ := json.Marshal(res)
		io.WriteString(w, string(strRes))
		return
	}
}
func checkCaptchaHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	defer exception(w)
	res := struct {
		Code int
		Msg  string
	}{
		0,
		"success",
	}

	captchaValue := r.FormValue("captchaValue")
	if "" == captchaValue {
		res.Code = 1
		res.Msg = "empty captchaValue!"
		strRes, _ := json.Marshal(res)
		io.WriteString(w, string(strRes))
		return
	}

	captchaId := r.FormValue("captchaId")
	if "" == captchaId {
		captchaCookie, err := r.Cookie("captchaId")
		if err != nil {
			res.Code = 4
			res.Msg = err.Error()
			strRes, _ := json.Marshal(res)
			io.WriteString(w, string(strRes))
			return
		}
		captchaId = captchaCookie.Value
	}
	if "" == captchaId {
		res.Code = 3
		res.Msg = "empty captchaId!"
		strRes, _ := json.Marshal(res)
		io.WriteString(w, string(strRes))
		return
	}

	if !captcha.VerifyString(captchaId, captchaValue) {
		res.Code = 2
		res.Msg = "check fail!"
	}

	strRes, _ := json.Marshal(res)
	io.WriteString(w, string(strRes))

}

func setStore() {
	conf, err := goconfig.ReadConfigFile(HomeRoot + "/conf/config.cfg")
	if err != nil {
		fmt.Println("error!")
		panic("error")
	}
	storeType, _ := conf.GetString("store", "type")
	host, _ := conf.GetString(storeType, "host")
	port, _ := conf.GetString(storeType, "port")
	expire, _ := conf.GetInt64(storeType, "expire")
	var maxIdle int64 = 10
	var maxActive int64 = 30
	if storeType == "redis" {
		maxIdle, _ = conf.GetInt64(storeType, "maxidle")
		maxActive, _ = conf.GetInt64(storeType, "maxactive")
	}

	var globalStore captcha.Store
	switch storeType {
	case "memcache":
		globalStore = captcha.GetMemcacheClient(host, port, expire)
	case "redis":
		globalStore = captcha.GetRedisPool(host, port, expire, maxIdle, maxActive)
	default:
		panic("not set store")
	}
	//globalStore := captcha.GetMemcacheVitessConnection()
	// globalStore := captcha.GetCouchBaseClient()
	captcha.SetCustomStore(globalStore)
}

func getCaptchaIdHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	d := struct {
		CaptchaId string
	}{
		captcha.New(),
	}
	res, err := json.Marshal(d)
	if err != nil {
		io.WriteString(w, "internal error!")
	}
	io.WriteString(w, string(res))

}

func generateCaptchaHandler(w http.ResponseWriter, r *http.Request) {
	//w.Header().Set("Content-Type", "application/json; charset=utf-8")
	imgWidth, err := strconv.Atoi(r.FormValue("width"))
	if err != nil {
		imgWidth = captcha.StdWidth
	}
	imgHeight, err := strconv.Atoi(r.FormValue("height"))
	if err != nil {
		imgHeight = captcha.StdHeight
	}
	if imgWidth == 0 || imgHeight == 0 {
		imgWidth = captcha.StdWidth
		imgHeight = captcha.StdHeight
	}
	fixColor, err := strconv.Atoi(r.FormValue("fixcolor"))
	var bFixColor bool
	if fixColor >= 1 {
		bFixColor = true
	} else {
		bFixColor = false
	}

	captchaId := captcha.New()
	cookie := new(http.Cookie)
	cookie.Name = "captchaId"
	cookie.Value = captchaId
	cookie.Path = "/"
	cookie.Domain = ".domain.com"

	http.SetCookie(w, cookie)
	captcha.WriteImage(w, captchaId, imgWidth, imgHeight, bFixColor)
	return
}

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	fmt.Println("init")
}

func main() {
	setStore()
	http.HandleFunc("/captcha/getid", getCaptchaIdHandler)
	http.HandleFunc("/captcha/generate", generateCaptchaHandler)
	http.HandleFunc("/captcha/process", processFormHandler)
	http.Handle("/captcha/image/", captcha.Server(captcha.StdWidth, captcha.StdHeight))
	//http.HandleFunc("/captcha/", showFormHandler)
	http.HandleFunc("/captcha/check", checkCaptchaHandler)
	fmt.Println("Server is at localhost:8666")
	if err := http.ListenAndServe(":8666", nil); err != nil {
		log.Fatal(err)
	}
}

const formTemplateSrc = `<!doctype html>
<head><title>Captcha Example</title></head>
<body>
<script>
function setSrcQuery(e, q) {
    var src  = e.src;
    var p = src.indexOf('?');
    if (p >= 0) {
        src = src.substr(0, p);
    }
    e.src = src + "?" + q
}

function playAudio() {
    var le = document.getElementById("lang");
    var lang = le.options[le.selectedIndex].value;
    var e = document.getElementById('audio')
    setSrcQuery(e, "lang=" + lang)
    e.style.display = 'block';
    e.autoplay = 'true';
    return false;
}

function changeLang() {
	var e = document.getElementById('audio')
	if (e.style.display == 'block') {
    	playAudio();
    }
}

function reload() {
    setSrcQuery(document.getElementById('image'), "reload=" + (new Date()).getTime());
    setSrcQuery(document.getElementById('audio'), (new Date()).getTime());
    return false;
}
</script>
<select id="lang" onchange="changeLang()">
<option value="en">English</option>
<option value="ru">Russian</option>
<option value="zh">Chinese</option>
</select>
<form action="/captcha/process" method=post>
<p>Type the numbers you see in the picture below:</p>
<p><img id=image src="/captcha/image/{{.CaptchaId}}.png" alt="Captcha image"></p>
<a href="#" onclick="reload()">Reload</a> | <a href="#" onclick="playAudio()">Play Audio</a>
<audio id=audio controls style="display:none" src="/captcha/{{.CaptchaId}}.wav" preload=none>
You browser doesn't support audio.
<a href="/captcha/download/{{.CaptchaId}}.wav">Download file</a> to play it in the external player.
</audio>
<input type=hidden name=captchaId value="{{.CaptchaId}}"><br>
<input name=captchaValue>
<input type=submit value=Submit>
</form>
`
