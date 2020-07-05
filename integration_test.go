package main_test

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ip-api/cache/batch"
	"github.com/ip-api/cache/cache"
	"github.com/ip-api/cache/fetcher"
	"github.com/ip-api/cache/field"
	"github.com/ip-api/cache/handlers"
	"github.com/ip-api/cache/structs"
	"github.com/ip-api/cache/util"
	"github.com/rs/zerolog"
	"github.com/valyala/fasthttp"
)

func TestSingle(t *testing.T) {
	t.Parallel()

	logger := zerolog.New(zerolog.ConsoleWriter{Out: util.ZerologTestWriter{T: t}, NoColor: true})

	cache := cache.New(1000000)
	client := &fetcher.Mock{}
	batches := batch.New(logger.With().Str("part", "batch").Logger(), cache, client)

	go batches.ProcessLoop()

	h := handlers.Handler{
		Logger:  logger.With().Str("part", "handler").Logger(),
		Batches: batches,
		Client:  client,
	}

	var ctx fasthttp.RequestCtx
	var req fasthttp.Request
	req.SetRequestURI("http://example.com/json/1.1.1.1?fields=8209")
	ctx.Init(&req, nil, nil)

	h.Index(&ctx)

	contentType := string(ctx.Response.Header.Peek(fasthttp.HeaderContentType))
	expectedContentType := "application/json"
	if contentType != expectedContentType {
		t.Errorf("expected %q got %q", expectedContentType, contentType)
	}

	body := string(ctx.Response.Body())
	expectedBody := `{"country":"Some Country","city":"Some City","query":"1.1.1.1"}`
	if body != expectedBody {
		t.Errorf("\nexpected\n%s\ngot\n%s", expectedBody, body)
	}
}

func TestBatch(t *testing.T) {
	t.Parallel()

	logger := zerolog.New(zerolog.ConsoleWriter{Out: util.ZerologTestWriter{T: t}, NoColor: true})

	cache := cache.New(1000000)
	client := &fetcher.Mock{}
	batches := batch.New(logger.With().Str("part", "batch").Logger(), cache, client)

	go batches.ProcessLoop()

	h := handlers.Handler{
		Logger:  logger.With().Str("part", "handler").Logger(),
		Batches: batches,
		Client:  client,
	}

	var ctx fasthttp.RequestCtx
	var req fasthttp.Request
	req.SetRequestURI("http://example.com/batch")
	req.SetBodyString(`
		[
			{
				"query": "1.1.1.1",
				"fields": 8209,
				"lang":"ja"
			},
			"2.2.2.2"
		  ]
	`)
	ctx.Init(&req, nil, nil)

	h.Index(&ctx)

	contentType := string(ctx.Response.Header.Peek(fasthttp.HeaderContentType))
	expectedContentType := "application/json"
	if contentType != expectedContentType {
		t.Errorf("expected %q got %q", expectedContentType, contentType)
	}

	body := string(ctx.Response.Body())
	expectedBody := `[{"country":"Some japanese Country","city":"Some japanese City","query":"1.1.1.1"},{"status":"success","country":"Some other Country","countryCode":"SO","region":"SX","regionName":"Some other Region","city":"Some other City","zip":"some other zip","lat":13,"lon":37,"timezone":"some/timezone","isp":"Some other ISP","org":"Some other Org","as":"Some other AS","query":"2.2.2.2"}]`
	if body != expectedBody {
		t.Errorf("\nexpected\n%s\ngot\n%s", expectedBody, body)
	}

	if len(client.Requests) != 1 {
		t.Errorf("expected 1 got %d", len(client.Requests))
	}

	if client.Requests[0] != 2 {
		t.Errorf("expected 2 got %d", client.Requests[0])
	}
}

func TestBatchBig(t *testing.T) {
	t.Parallel()

	logger := zerolog.New(zerolog.ConsoleWriter{Out: util.ZerologTestWriter{T: t}, NoColor: true})

	cache := cache.New(1000000)
	client := &fetcher.Mock{}
	batches := batch.New(logger.With().Str("part", "batch").Logger(), cache, client)

	go batches.ProcessLoop()

	h := handlers.Handler{
		Logger:  logger.With().Str("part", "handler").Logger(),
		Batches: batches,
		Client:  client,
	}

	var ctx fasthttp.RequestCtx
	var req fasthttp.Request
	req.SetRequestURI("http://example.com/batch")

	// 150 IPs.
	req.SetBodyString(`
		[
			"0.0.0.0","1.1.1.1","2.2.2.2","3.3.3.3","4.4.4.4","5.5.5.5","6.6.6.6","7.7.7.7","8.8.8.8","9.9.9.9",
			"10.10.10.10","11.11.11.11","12.12.12.12","13.13.13.13","14.14.14.14","15.15.15.15","16.16.16.16","17.17.17.17",
			"18.18.18.18","19.19.19.19","20.20.20.20","21.21.21.21","22.22.22.22","23.23.23.23","24.24.24.24","25.25.25.25",
			"26.26.26.26","27.27.27.27","28.28.28.28","29.29.29.29","30.30.30.30","31.31.31.31","32.32.32.32","33.33.33.33",
			"34.34.34.34","35.35.35.35","36.36.36.36","37.37.37.37","38.38.38.38","39.39.39.39","40.40.40.40","41.41.41.41",
			"42.42.42.42","43.43.43.43","44.44.44.44","45.45.45.45","46.46.46.46","47.47.47.47","48.48.48.48","49.49.49.49",
			"50.50.50.50","51.51.51.51","52.52.52.52","53.53.53.53","54.54.54.54","55.55.55.55","56.56.56.56","57.57.57.57",
			"58.58.58.58","59.59.59.59","60.60.60.60","61.61.61.61","62.62.62.62","63.63.63.63","64.64.64.64","65.65.65.65",
			"66.66.66.66","67.67.67.67","68.68.68.68","69.69.69.69","70.70.70.70","71.71.71.71","72.72.72.72","73.73.73.73",
			"74.74.74.74","75.75.75.75","76.76.76.76","77.77.77.77","78.78.78.78","79.79.79.79","80.80.80.80","81.81.81.81",
			"82.82.82.82","83.83.83.83","84.84.84.84","85.85.85.85","86.86.86.86","87.87.87.87","88.88.88.88","89.89.89.89",
			"90.90.90.90","91.91.91.91","92.92.92.92","93.93.93.93","94.94.94.94","95.95.95.95","96.96.96.96","97.97.97.97",
			"98.98.98.98","99.99.99.99","100.100.100.100","101.101.101.101","102.102.102.102","103.103.103.103","104.104.104.104",
			"105.105.105.105","106.106.106.106","107.107.107.107","108.108.108.108","109.109.109.109","110.110.110.110",
			"111.111.111.111","112.112.112.112","113.113.113.113","114.114.114.114","115.115.115.115","116.116.116.116",
			"117.117.117.117","118.118.118.118","119.119.119.119","120.120.120.120","121.121.121.121","122.122.122.122",
			"123.123.123.123","124.124.124.124","125.125.125.125","126.126.126.126","127.127.127.127","128.128.128.128",
			"129.129.129.129","130.130.130.130","131.131.131.131","132.132.132.132","133.133.133.133","134.134.134.134",
			"135.135.135.135","136.136.136.136","137.137.137.137","138.138.138.138","139.139.139.139","140.140.140.140",
			"141.141.141.141","142.142.142.142","143.143.143.143","144.144.144.144","145.145.145.145","146.146.146.146",
			"147.147.147.147","148.148.148.148","149.149.149.149"
		]
	`)
	ctx.Init(&req, nil, nil)

	h.Index(&ctx)

	contentType := string(ctx.Response.Header.Peek(fasthttp.HeaderContentType))
	expectedContentType := "application/json"
	if contentType != expectedContentType {
		t.Errorf("expected %q got %q", expectedContentType, contentType)
	}

	body := string(ctx.Response.Body())
	expectedBody := `[{"country":"0.0.0.0en","city":"0.0.0.0en","query":"0.0.0.0"},{"country":"Some Country","city":"Some City","query":"1.1.1.1"},{"status":"success","country":"Some other Country","countryCode":"SO","region":"SX","regionName":"Some other Region","city":"Some other City","zip":"some other zip","lat":13,"lon":37,"timezone":"some/timezone","isp":"Some other ISP","org":"Some other Org","as":"Some other AS","query":"2.2.2.2"},{"country":"3.3.3.3en","city":"3.3.3.3en","query":"3.3.3.3"},{"country":"4.4.4.4en","city":"4.4.4.4en","query":"4.4.4.4"},{"country":"5.5.5.5en","city":"5.5.5.5en","query":"5.5.5.5"},{"country":"6.6.6.6en","city":"6.6.6.6en","query":"6.6.6.6"},{"country":"7.7.7.7en","city":"7.7.7.7en","query":"7.7.7.7"},{"country":"8.8.8.8en","city":"8.8.8.8en","query":"8.8.8.8"},{"country":"9.9.9.9en","city":"9.9.9.9en","query":"9.9.9.9"},{"country":"10.10.10.10en","city":"10.10.10.10en","query":"10.10.10.10"},{"country":"11.11.11.11en","city":"11.11.11.11en","query":"11.11.11.11"},{"country":"12.12.12.12en","city":"12.12.12.12en","query":"12.12.12.12"},{"country":"13.13.13.13en","city":"13.13.13.13en","query":"13.13.13.13"},{"country":"14.14.14.14en","city":"14.14.14.14en","query":"14.14.14.14"},{"country":"15.15.15.15en","city":"15.15.15.15en","query":"15.15.15.15"},{"country":"16.16.16.16en","city":"16.16.16.16en","query":"16.16.16.16"},{"country":"17.17.17.17en","city":"17.17.17.17en","query":"17.17.17.17"},{"country":"18.18.18.18en","city":"18.18.18.18en","query":"18.18.18.18"},{"country":"19.19.19.19en","city":"19.19.19.19en","query":"19.19.19.19"},{"country":"20.20.20.20en","city":"20.20.20.20en","query":"20.20.20.20"},{"country":"21.21.21.21en","city":"21.21.21.21en","query":"21.21.21.21"},{"country":"22.22.22.22en","city":"22.22.22.22en","query":"22.22.22.22"},{"country":"23.23.23.23en","city":"23.23.23.23en","query":"23.23.23.23"},{"country":"24.24.24.24en","city":"24.24.24.24en","query":"24.24.24.24"},{"country":"25.25.25.25en","city":"25.25.25.25en","query":"25.25.25.25"},{"country":"26.26.26.26en","city":"26.26.26.26en","query":"26.26.26.26"},{"country":"27.27.27.27en","city":"27.27.27.27en","query":"27.27.27.27"},{"country":"28.28.28.28en","city":"28.28.28.28en","query":"28.28.28.28"},{"country":"29.29.29.29en","city":"29.29.29.29en","query":"29.29.29.29"},{"country":"30.30.30.30en","city":"30.30.30.30en","query":"30.30.30.30"},{"country":"31.31.31.31en","city":"31.31.31.31en","query":"31.31.31.31"},{"country":"32.32.32.32en","city":"32.32.32.32en","query":"32.32.32.32"},{"country":"33.33.33.33en","city":"33.33.33.33en","query":"33.33.33.33"},{"country":"34.34.34.34en","city":"34.34.34.34en","query":"34.34.34.34"},{"country":"35.35.35.35en","city":"35.35.35.35en","query":"35.35.35.35"},{"country":"36.36.36.36en","city":"36.36.36.36en","query":"36.36.36.36"},{"country":"37.37.37.37en","city":"37.37.37.37en","query":"37.37.37.37"},{"country":"38.38.38.38en","city":"38.38.38.38en","query":"38.38.38.38"},{"country":"39.39.39.39en","city":"39.39.39.39en","query":"39.39.39.39"},{"country":"40.40.40.40en","city":"40.40.40.40en","query":"40.40.40.40"},{"country":"41.41.41.41en","city":"41.41.41.41en","query":"41.41.41.41"},{"country":"42.42.42.42en","city":"42.42.42.42en","query":"42.42.42.42"},{"country":"43.43.43.43en","city":"43.43.43.43en","query":"43.43.43.43"},{"country":"44.44.44.44en","city":"44.44.44.44en","query":"44.44.44.44"},{"country":"45.45.45.45en","city":"45.45.45.45en","query":"45.45.45.45"},{"country":"46.46.46.46en","city":"46.46.46.46en","query":"46.46.46.46"},{"country":"47.47.47.47en","city":"47.47.47.47en","query":"47.47.47.47"},{"country":"48.48.48.48en","city":"48.48.48.48en","query":"48.48.48.48"},{"country":"49.49.49.49en","city":"49.49.49.49en","query":"49.49.49.49"},{"country":"50.50.50.50en","city":"50.50.50.50en","query":"50.50.50.50"},{"country":"51.51.51.51en","city":"51.51.51.51en","query":"51.51.51.51"},{"country":"52.52.52.52en","city":"52.52.52.52en","query":"52.52.52.52"},{"country":"53.53.53.53en","city":"53.53.53.53en","query":"53.53.53.53"},{"country":"54.54.54.54en","city":"54.54.54.54en","query":"54.54.54.54"},{"country":"55.55.55.55en","city":"55.55.55.55en","query":"55.55.55.55"},{"country":"56.56.56.56en","city":"56.56.56.56en","query":"56.56.56.56"},{"country":"57.57.57.57en","city":"57.57.57.57en","query":"57.57.57.57"},{"country":"58.58.58.58en","city":"58.58.58.58en","query":"58.58.58.58"},{"country":"59.59.59.59en","city":"59.59.59.59en","query":"59.59.59.59"},{"country":"60.60.60.60en","city":"60.60.60.60en","query":"60.60.60.60"},{"country":"61.61.61.61en","city":"61.61.61.61en","query":"61.61.61.61"},{"country":"62.62.62.62en","city":"62.62.62.62en","query":"62.62.62.62"},{"country":"63.63.63.63en","city":"63.63.63.63en","query":"63.63.63.63"},{"country":"64.64.64.64en","city":"64.64.64.64en","query":"64.64.64.64"},{"country":"65.65.65.65en","city":"65.65.65.65en","query":"65.65.65.65"},{"country":"66.66.66.66en","city":"66.66.66.66en","query":"66.66.66.66"},{"country":"67.67.67.67en","city":"67.67.67.67en","query":"67.67.67.67"},{"country":"68.68.68.68en","city":"68.68.68.68en","query":"68.68.68.68"},{"country":"69.69.69.69en","city":"69.69.69.69en","query":"69.69.69.69"},{"country":"70.70.70.70en","city":"70.70.70.70en","query":"70.70.70.70"},{"country":"71.71.71.71en","city":"71.71.71.71en","query":"71.71.71.71"},{"country":"72.72.72.72en","city":"72.72.72.72en","query":"72.72.72.72"},{"country":"73.73.73.73en","city":"73.73.73.73en","query":"73.73.73.73"},{"country":"74.74.74.74en","city":"74.74.74.74en","query":"74.74.74.74"},{"country":"75.75.75.75en","city":"75.75.75.75en","query":"75.75.75.75"},{"country":"76.76.76.76en","city":"76.76.76.76en","query":"76.76.76.76"},{"country":"77.77.77.77en","city":"77.77.77.77en","query":"77.77.77.77"},{"country":"78.78.78.78en","city":"78.78.78.78en","query":"78.78.78.78"},{"country":"79.79.79.79en","city":"79.79.79.79en","query":"79.79.79.79"},{"country":"80.80.80.80en","city":"80.80.80.80en","query":"80.80.80.80"},{"country":"81.81.81.81en","city":"81.81.81.81en","query":"81.81.81.81"},{"country":"82.82.82.82en","city":"82.82.82.82en","query":"82.82.82.82"},{"country":"83.83.83.83en","city":"83.83.83.83en","query":"83.83.83.83"},{"country":"84.84.84.84en","city":"84.84.84.84en","query":"84.84.84.84"},{"country":"85.85.85.85en","city":"85.85.85.85en","query":"85.85.85.85"},{"country":"86.86.86.86en","city":"86.86.86.86en","query":"86.86.86.86"},{"country":"87.87.87.87en","city":"87.87.87.87en","query":"87.87.87.87"},{"country":"88.88.88.88en","city":"88.88.88.88en","query":"88.88.88.88"},{"country":"89.89.89.89en","city":"89.89.89.89en","query":"89.89.89.89"},{"country":"90.90.90.90en","city":"90.90.90.90en","query":"90.90.90.90"},{"country":"91.91.91.91en","city":"91.91.91.91en","query":"91.91.91.91"},{"country":"92.92.92.92en","city":"92.92.92.92en","query":"92.92.92.92"},{"country":"93.93.93.93en","city":"93.93.93.93en","query":"93.93.93.93"},{"country":"94.94.94.94en","city":"94.94.94.94en","query":"94.94.94.94"},{"country":"95.95.95.95en","city":"95.95.95.95en","query":"95.95.95.95"},{"country":"96.96.96.96en","city":"96.96.96.96en","query":"96.96.96.96"},{"country":"97.97.97.97en","city":"97.97.97.97en","query":"97.97.97.97"},{"country":"98.98.98.98en","city":"98.98.98.98en","query":"98.98.98.98"},{"country":"99.99.99.99en","city":"99.99.99.99en","query":"99.99.99.99"},{"country":"100.100.100.100en","city":"100.100.100.100en","query":"100.100.100.100"},{"country":"101.101.101.101en","city":"101.101.101.101en","query":"101.101.101.101"},{"country":"102.102.102.102en","city":"102.102.102.102en","query":"102.102.102.102"},{"country":"103.103.103.103en","city":"103.103.103.103en","query":"103.103.103.103"},{"country":"104.104.104.104en","city":"104.104.104.104en","query":"104.104.104.104"},{"country":"105.105.105.105en","city":"105.105.105.105en","query":"105.105.105.105"},{"country":"106.106.106.106en","city":"106.106.106.106en","query":"106.106.106.106"},{"country":"107.107.107.107en","city":"107.107.107.107en","query":"107.107.107.107"},{"country":"108.108.108.108en","city":"108.108.108.108en","query":"108.108.108.108"},{"country":"109.109.109.109en","city":"109.109.109.109en","query":"109.109.109.109"},{"country":"110.110.110.110en","city":"110.110.110.110en","query":"110.110.110.110"},{"country":"111.111.111.111en","city":"111.111.111.111en","query":"111.111.111.111"},{"country":"112.112.112.112en","city":"112.112.112.112en","query":"112.112.112.112"},{"country":"113.113.113.113en","city":"113.113.113.113en","query":"113.113.113.113"},{"country":"114.114.114.114en","city":"114.114.114.114en","query":"114.114.114.114"},{"country":"115.115.115.115en","city":"115.115.115.115en","query":"115.115.115.115"},{"country":"116.116.116.116en","city":"116.116.116.116en","query":"116.116.116.116"},{"country":"117.117.117.117en","city":"117.117.117.117en","query":"117.117.117.117"},{"country":"118.118.118.118en","city":"118.118.118.118en","query":"118.118.118.118"},{"country":"119.119.119.119en","city":"119.119.119.119en","query":"119.119.119.119"},{"country":"120.120.120.120en","city":"120.120.120.120en","query":"120.120.120.120"},{"country":"121.121.121.121en","city":"121.121.121.121en","query":"121.121.121.121"},{"country":"122.122.122.122en","city":"122.122.122.122en","query":"122.122.122.122"},{"country":"123.123.123.123en","city":"123.123.123.123en","query":"123.123.123.123"},{"country":"124.124.124.124en","city":"124.124.124.124en","query":"124.124.124.124"},{"country":"125.125.125.125en","city":"125.125.125.125en","query":"125.125.125.125"},{"country":"126.126.126.126en","city":"126.126.126.126en","query":"126.126.126.126"},{"country":"127.127.127.127en","city":"127.127.127.127en","query":"127.127.127.127"},{"country":"128.128.128.128en","city":"128.128.128.128en","query":"128.128.128.128"},{"country":"129.129.129.129en","city":"129.129.129.129en","query":"129.129.129.129"},{"country":"130.130.130.130en","city":"130.130.130.130en","query":"130.130.130.130"},{"country":"131.131.131.131en","city":"131.131.131.131en","query":"131.131.131.131"},{"country":"132.132.132.132en","city":"132.132.132.132en","query":"132.132.132.132"},{"country":"133.133.133.133en","city":"133.133.133.133en","query":"133.133.133.133"},{"country":"134.134.134.134en","city":"134.134.134.134en","query":"134.134.134.134"},{"country":"135.135.135.135en","city":"135.135.135.135en","query":"135.135.135.135"},{"country":"136.136.136.136en","city":"136.136.136.136en","query":"136.136.136.136"},{"country":"137.137.137.137en","city":"137.137.137.137en","query":"137.137.137.137"},{"country":"138.138.138.138en","city":"138.138.138.138en","query":"138.138.138.138"},{"country":"139.139.139.139en","city":"139.139.139.139en","query":"139.139.139.139"},{"country":"140.140.140.140en","city":"140.140.140.140en","query":"140.140.140.140"},{"country":"141.141.141.141en","city":"141.141.141.141en","query":"141.141.141.141"},{"country":"142.142.142.142en","city":"142.142.142.142en","query":"142.142.142.142"},{"country":"143.143.143.143en","city":"143.143.143.143en","query":"143.143.143.143"},{"country":"144.144.144.144en","city":"144.144.144.144en","query":"144.144.144.144"},{"country":"145.145.145.145en","city":"145.145.145.145en","query":"145.145.145.145"},{"country":"146.146.146.146en","city":"146.146.146.146en","query":"146.146.146.146"},{"country":"147.147.147.147en","city":"147.147.147.147en","query":"147.147.147.147"},{"country":"148.148.148.148en","city":"148.148.148.148en","query":"148.148.148.148"},{"country":"149.149.149.149en","city":"149.149.149.149en","query":"149.149.149.149"}]`
	if body != expectedBody {
		t.Errorf("\nexpected\n%s\ngot\n%s", expectedBody, body)
	}

	// The batches could be split multiple ways depending on the timing of the batch processing so don't test this here.
}

func TestCache(t *testing.T) {
	logger := zerolog.New(zerolog.ConsoleWriter{Out: util.ZerologTestWriter{T: t}, NoColor: true})

	cache := cache.New(1000000)
	client := &fetcher.Mock{}
	batches := batch.New(logger.With().Str("part", "batch").Logger(), cache, client)

	go batches.ProcessLoop()

	h := handlers.Handler{
		Logger:  logger.With().Str("part", "handler").Logger(),
		Batches: batches,
		Client:  client,
	}

	currentTime := time.Now()
	util.Now = func() time.Time {
		return currentTime
	}
	defer func() {
		util.Now = time.Now
	}()

	for i := 0; i < 3; i++ {
		var ctx fasthttp.RequestCtx
		var req fasthttp.Request
		req.SetRequestURI("http://example.com/json/1.1.1.1?fields=" + strconv.Itoa(int(field.FromCSV("country,city,query"))))
		ctx.Init(&req, nil, nil)

		h.Index(&ctx)

		body := string(ctx.Response.Body())
		expectedBody := `{"country":"Some Country","city":"Some City","query":"1.1.1.1"}`
		if body != expectedBody {
			t.Errorf("\nexpected\n%s\ngot\n%s", expectedBody, body)
		}

		// Forward time by 40 seconds.
		// Our mock fetcher caches for one minute, so this means the second request should be cached,
		// but the third request shouldn't.
		currentTime = currentTime.Add(time.Second * 40)
	}

	// Expect 2 requests as the middle one was cached.
	if len(client.Requests) != 2 {
		t.Errorf("expected 2 got %d", len(client.Requests))
	}
}

func TestHammer(t *testing.T) {
	logger := zerolog.New(zerolog.ConsoleWriter{Out: util.ZerologTestWriter{T: t}, NoColor: true})

	cache := cache.New(1000000)
	client := &fetcher.Mock{}
	batches := batch.New(logger.With().Str("part", "batch").Logger(), cache, client)

	go batches.ProcessLoop()

	h := handlers.Handler{
		Logger:  logger.With().Str("part", "handler").Logger(),
		Batches: batches,
		Client:  client,
	}

	var wg sync.WaitGroup

	// Have 10 goroutines each send 1000 batch requests with each 1-20 IPs.
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for i := 0; i < 100; i++ {
				ips := make([]string, 0, 1+rand.Intn(20))
				responses := make(structs.Responses, 0, cap(ips))
				for j := 0; j < cap(ips); j++ {
					ip := fmt.Sprintf("1.1.%d.%d", rand.Intn(10), rand.Intn(256))
					ips = append(ips, ip)

					responses = append(responses, fetcher.MockResponseFor(ip+"en"))
				}

				var ctx fasthttp.RequestCtx
				var req fasthttp.Request
				req.SetRequestURI("http://example.com/batch")
				req.SetBodyString(`["` + strings.Join(ips, `","`) + `"]`)
				ctx.Init(&req, nil, nil)

				h.Index(&ctx)

				body := string(ctx.Response.Body())
				expectedBody, _ := responses.MarshalJSON()
				if body != string(expectedBody) {
					t.Errorf("\nexpected\n%s\ngot\n%s", expectedBody, body)
				}

				time.Sleep(time.Duration(rand.Intn(10)) * time.Millisecond)
			}
		}()
	}

	for i := 0; i < 30; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for i := 0; i < 100; i++ {
				ip := fmt.Sprintf("1.1.%d.%d", rand.Intn(10), rand.Intn(256))
				response := fetcher.MockResponseFor(ip + "en").Trim(8209)

				var ctx fasthttp.RequestCtx
				var req fasthttp.Request
				req.SetRequestURI("http://example.com/json/" + ip + "?fields=8209")
				ctx.Init(&req, nil, nil)

				h.Index(&ctx)

				body := string(ctx.Response.Body())
				expectedBody, _ := response.MarshalJSON()
				if body != string(expectedBody) {
					t.Errorf("\nexpected\n%s\ngot\n%s", expectedBody, body)
				}

				time.Sleep(time.Duration(rand.Intn(10)) * time.Millisecond)
			}
		}()
	}

	wg.Wait()
}
