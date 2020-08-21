package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"strconv"

	"github.com/mailru/easyjson"
	"github.com/mailru/easyjson/jwriter"
	"github.com/rs/zerolog"
	"github.com/valyala/fasthttp"

	"github.com/ip-api/proxy/internal/batch"
	"github.com/ip-api/proxy/internal/cache"
	"github.com/ip-api/proxy/internal/fetcher"
	"github.com/ip-api/proxy/internal/field"
	"github.com/ip-api/proxy/internal/structs"
	"github.com/ip-api/proxy/internal/util"
	"github.com/ip-api/proxy/internal/wait"
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
	strSlashDebug                             = []byte("/debug")
	strSlashJson                              = []byte("/json")
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

const defaultLanguage = "en"

type Handler struct {
	Logger  zerolog.Logger
	Cache   *cache.Cache
	Batches *batch.Batches
	Client  fetcher.Client
}

func (h Handler) writeResponse(ctx *fasthttp.RequestCtx, response easyjson.Marshaler) {
	jw := &jwriter.Writer{}
	response.MarshalEasyJSON(jw)
	if _, err := jw.DumpTo(ctx); err != nil {
		h.Logger.Error().Err(err).Msg("failed to write response")
	}
}

// /json/{query}
//   ?fields=<bitmap | comma separated list>
//   ?lang=<lang>
//   ?callback=<function name>
func (h Handler) single(ctx *fasthttp.RequestCtx) {
	path := ctx.Path()
	qa := ctx.QueryArgs()

	var fields field.Fields
	fieldsStr := util.B2s(qa.Peek("fields"))
	if len(fieldsStr) == 0 {
		fields = field.Default
	} else if i, err := strconv.Atoi(fieldsStr); err == nil {
		fields = field.FromInt(i)
	} else {
		fields = field.FromCSV(fieldsStr)
	}

	lang := string(qa.Peek("lang"))
	if lang == "" {
		lang = defaultLanguage
	} else if _, ok := languages[lang]; !ok {
		h.writeResponse(ctx, structs.ErrorResponse("fail", "invalid language").Trim(fields))
		return
	}

	if len(path) <= len(strSlashJsonSlash) {
		r, err := h.Client.FetchSelf(lang, fields)
		if err != nil {
			h.writeResponse(ctx, structs.ErrorResponse("fail", "error in upstream").Trim(fields))
			return
		}

		h.writeResponse(ctx, r.Trim(fields))
		return
	}

	ip := string(path[len(strSlashJsonSlash):])

	if net.ParseIP(ip) == nil {
		h.writeResponse(ctx, structs.ErrorResponse("fail", "invalid query").Trim(fields))
		return
	}

	entry, c := h.Batches.Add(ip, lang, fields)

	if c != nil {
		// Wait for the entry to contain valid data.
		<-c
	}

	h.writeResponse(ctx, entry.Response.Trim(fields))
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
	qa := ctx.QueryArgs()

	var defaultFields field.Fields
	if fieldsStr := util.B2s(qa.Peek("fields")); len(fieldsStr) == 0 {
		defaultFields = field.Default
	} else if i, err := strconv.Atoi(fieldsStr); err == nil {
		defaultFields = field.FromInt(i)
	} else {
		defaultFields = field.FromCSV(fieldsStr)
	}

	var body []interface{}
	if err := json.Unmarshal(ctx.PostBody(), &body); err != nil {
		h.writeResponse(ctx, structs.Responses{
			structs.ErrorResponse("fail", "invalid body").Trim(defaultFields),
		})
		return
	}

	defaultLang := string(qa.Peek("lang"))
	if defaultLang == "" {
		defaultLang = defaultLanguage
	} else if _, ok := languages[defaultLang]; !ok {
		h.writeResponse(ctx, structs.Responses{
			structs.ErrorResponse("fail", "invalid language").Trim(defaultFields),
		})
		return
	}

	fields := make([]field.Fields, len(body))
	entries := make([]*structs.CacheEntry, len(body))
	w := wait.New()

	for i, part := range body {
		var ip string
		var lang string

		if ipStr, ok := part.(string); ok {
			if net.ParseIP(ipStr) == nil {
				fields[i] = defaultFields
				entries[i] = &structs.CacheEntry{
					Response: structs.ErrorResponse("fail", "invalid query"),
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
						Response: structs.ErrorResponse("fail", "invalid query"),
					}
					continue
				}

				if net.ParseIP(ip) == nil {
					entries[i] = &structs.CacheEntry{
						Response: structs.ErrorResponse("fail", "invalid query"),
					}
					continue
				}
			} else {
				entries[i] = &structs.CacheEntry{
					Response: structs.ErrorResponse("fail", "invalid query"),
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

	h.writeResponse(ctx, responses)
}

func (h Handler) debug(ctx *fasthttp.RequestCtx) {
	if err := json.NewEncoder(ctx).Encode(map[string]interface{}{
		"fetcher": h.Client.Debug(),
	}); err != nil {
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

	if bytes.HasPrefix(path, strSlashJsonSlash) || bytes.Equal(path, strSlashJson) {
		h.single(ctx)
	} else if bytes.Equal(path, strSlashBatch) {
		h.batch(ctx)
	} else if bytes.Equal(path, strSlashDebug) {
		h.debug(ctx)
	} else {
		ctx.Response.SetStatusCode(fasthttp.StatusNotFound)
	}
}
