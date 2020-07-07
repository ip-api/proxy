package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"strconv"

	"github.com/ip-api/proxy/batch"
	"github.com/ip-api/proxy/cache"
	"github.com/ip-api/proxy/fetcher"
	"github.com/ip-api/proxy/field"
	"github.com/ip-api/proxy/structs"
	"github.com/ip-api/proxy/util"
	"github.com/ip-api/proxy/wait"
	"github.com/mailru/easyjson/jwriter"
	"github.com/rs/zerolog"
	"github.com/valyala/fasthttp"
)

var (
	strAccessControlAllowHeaders              = []byte("Access-Control-Allow-Headers")
	strAccessControlAllowMethods              = []byte("Access-Control-Allow-Methods")
	strAccessControlAllowOrigin               = []byte("Access-Control-Allow-Origin")
	strApplicationJson                        = []byte("application/json")
	strCacheControl                           = []byte("Cache-Control")
	strContentType                            = []byte("Content-Type")
	strContentTypeContentLengthAcceptEncoding = []byte("Content-Type, Content-Length, Accept-Encoding")
	strOPTIONS                                = []byte("OPTIONS")
	strPostGetOptions                         = []byte("POST, GET, OPTIONS")
	strSlashBatch                             = []byte("/batch")
	strSlashJsonSlash                         = []byte("/json/")
	strStar                                   = []byte("*")
	strYesEverything                          = []byte("public, max-age=1800")
)

var languages = map[string]struct{}{
	"en":    {},
	"de":    {},
	"es":    {},
	"pt-BR": {},
	"fr":    {},
	"ja":    {},
	"zh-CN": {},
	"ru":    {},
}

var (
	fail         = "fail"
	invalidQuery = "invalid query"
)

const defaultLanguage = "en"

type Handler struct {
	Logger  zerolog.Logger
	Cache   *cache.Cache
	Batches *batch.Batches
	Client  fetcher.Client
}

// /json/{query}
//   ?fields=<bitmap | comma separated list>
//   ?lang=<lang>
//   ?callback=<function name>
func (h Handler) single(ctx *fasthttp.RequestCtx) {
	path := ctx.Path()
	qa := ctx.QueryArgs()

	lang := string(qa.Peek("lang"))
	if lang == "" {
		lang = defaultLanguage
	} else if _, ok := languages[lang]; !ok {
		ctx.SetBodyString(`{"message":"invalid language","status":"fail"}`)
		ctx.Response.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}

	var fields field.Fields
	fieldsStr := util.B2s(qa.Peek("fields"))
	if len(fieldsStr) == 0 {
		fields = field.Default
	} else if i, err := strconv.Atoi(fieldsStr); err == nil {
		fields = field.FromInt(i)
	} else {
		fields = field.FromCSV(fieldsStr)
	}

	ip := string(path[len(strSlashJsonSlash):])
	if len(ip) == 0 {
		r, err := h.Client.FetchSelf(lang)
		if err != nil {
			ctx.SetBodyString(`{"message":"error in upstream","status":"fail"}`)
			ctx.Response.SetStatusCode(fasthttp.StatusInternalServerError)
			return
		}

		jw := &jwriter.Writer{}
		r.Trim(fields).MarshalEasyJSON(jw)
		if _, err := jw.DumpTo(ctx); err != nil {
			h.Logger.Error().Err(err).Msg("failed to write response")
		}
		return
	}

	if net.ParseIP(ip) == nil {
		ctx.SetBodyString(`{"message":"invalid query","status":"fail"}`)
		ctx.Response.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}

	entry, c := h.Batches.Add(ip, lang, fields)

	if c != nil {
		// Wait for the entry to contain valid data.
		<-c
	}

	jw := &jwriter.Writer{}
	entry.Response.Trim(fields).MarshalEasyJSON(jw)
	if _, err := jw.DumpTo(ctx); err != nil {
		h.Logger.Error().Err(err).Msg("failed to write response")
	}
}

// /batch
//   ?fields=<bitmap | comma separated list>
// ["1.1.1.1"|
// {
//   "query": "IPv4/IPv6 required",
//   "fields": "response fields optional",
//   "lang": "response language optional"
// }
// ]
func (h Handler) batch(ctx *fasthttp.RequestCtx) {
	w := wait.New()

	var body []interface{}
	if err := json.Unmarshal(ctx.PostBody(), &body); err != nil {
		ctx.SetBodyString(`{"message":"invalid request","status":"fail"}`)
		ctx.Response.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}

	qa := ctx.QueryArgs()

	defaultLang := string(qa.Peek("lang"))
	if defaultLang == "" {
		defaultLang = defaultLanguage
	} else if _, ok := languages[defaultLang]; !ok {
		ctx.SetBodyString(`{"message":"invalid language","status":"fail"}`)
		ctx.Response.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}

	var defaultFields field.Fields
	if fieldsStr := util.B2s(qa.Peek("fields")); len(fieldsStr) == 0 {
		defaultFields = field.Default
	} else if i, err := strconv.Atoi(fieldsStr); err == nil {
		defaultFields = field.FromInt(i)
	} else {
		defaultFields = field.FromCSV(fieldsStr)
	}

	fields := make([]field.Fields, len(body))
	entries := make([]*structs.CacheEntry, len(body))

	for i, part := range body {
		var ip string
		var lang string

		if ipStr, ok := part.(string); ok {
			if net.ParseIP(ipStr) == nil {
				fields[i] = defaultFields
				entries[i] = &structs.CacheEntry{
					Response: structs.Response{
						Status:  &fail,
						Message: &invalidQuery,
						Query:   &ipStr,
					},
				}
				continue
			} else {
				ip = ipStr
				lang = defaultLang
				fields[i] = defaultFields
			}
		} else if m, ok := part.(map[string]interface{}); ok {
			if fieldsIf, ok := m["fields"]; ok {
				if s, ok := fieldsIf.(string); ok {
					if n, err := strconv.Atoi(s); err == nil {
						fields[i] = field.FromInt(n)
					} else {
						fields[i] = field.FromCSV(s)
					}
				} else if f, ok := fieldsIf.(float64); ok {
					fields[i] = field.FromInt(int(f))
				} else {
					fields[i] = defaultFields
				}
			} else {
				fields[i] = defaultFields
			}

			if ipIf, ok := m["query"]; ok {
				ip, ok = ipIf.(string)
				if !ok {
					entries[i] = &structs.CacheEntry{
						Response: structs.Response{
							Status:  &fail,
							Message: &invalidQuery,
						},
					}
					continue
				}

				if net.ParseIP(ip) == nil {
					entries[i] = &structs.CacheEntry{
						Response: structs.Response{
							Status:  &fail,
							Message: &invalidQuery,
							Query:   &ip,
						},
					}
					continue
				}
			} else {
				entries[i] = &structs.CacheEntry{
					Response: structs.Response{
						Status:  &fail,
						Message: &invalidQuery,
					},
				}
				continue
			}

			if langIf, ok := m["lang"]; ok {
				lang, ok = langIf.(string)
				if !ok {
					lang = defaultLang
				} else if _, ok := languages[lang]; !ok {
					lang = defaultLang
				}
			} else {
				lang = defaultLang
			}
		}

		entry, c := h.Batches.Add(ip, lang, fields[i])
		entries[i] = entry
		if c != nil {
			w.Add(c)
		}
	}

	w.Wait()

	responses := make(structs.Responses, 0, len(entries))
	for i, e := range entries {
		responses = append(responses, e.Response.Trim(fields[i]))
	}

	jw := &jwriter.Writer{}
	responses.MarshalEasyJSON(jw)
	if _, err := jw.DumpTo(ctx); err != nil {
		h.Logger.Error().Err(err).Msg("failed to write responses")
	}
}

func (h Handler) Index(ctx *fasthttp.RequestCtx) {
	defer func() {
		if err := recover(); err != nil {
			h.Logger.Error().Err(fmt.Errorf("%v", err)).Msg("panic")
		}
	}()

	ctx.Response.Header.SetCanonical(strCacheControl, strYesEverything)
	ctx.Response.Header.SetCanonical(strAccessControlAllowOrigin, strStar)
	ctx.Response.Header.SetCanonical(strAccessControlAllowMethods, strPostGetOptions)
	ctx.Response.Header.SetCanonical(strAccessControlAllowHeaders, strContentTypeContentLengthAcceptEncoding)
	ctx.Response.Header.SetCanonical(strContentType, strApplicationJson)

	if bytes.Equal(ctx.Method(), strOPTIONS) {
		ctx.Response.SetStatusCode(fasthttp.StatusOK)
		return
	}

	path := ctx.Path()

	if bytes.HasPrefix(path, strSlashJsonSlash) {
		h.single(ctx)
	} else if bytes.HasPrefix(path, strSlashBatch) {
		h.batch(ctx)
	} else {
		ctx.Response.SetStatusCode(fasthttp.StatusNotFound)
	}
}
