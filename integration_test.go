package main_test

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ip-api/proxy/batch"
	"github.com/ip-api/proxy/cache"
	"github.com/ip-api/proxy/fetcher"
	"github.com/ip-api/proxy/field"
	"github.com/ip-api/proxy/handlers"
	"github.com/ip-api/proxy/structs"
	"github.com/ip-api/proxy/util"
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
			"3.3.3.3",
			"2.2.2.2",
			{
				"query": "1.1.1.1",
				"fields": "country",
				"lang": "ja"
			}
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
	expectedBody := `[{"status":"","country":"3.3.3.3en","countryCode":"","region":"","regionName":"","city":"3.3.3.3en","zip":"","lat":0,"lon":0,"timezone":"","isp":"","org":"","as":"","message":"","query":"3.3.3.3"},{"status":"success","country":"Some other Country","countryCode":"SO","region":"SX","regionName":"Some other Region","city":"Some other City","zip":"some other zip","lat":13,"lon":37,"timezone":"some/timezone","isp":"Some other ISP","org":"Some other Org","as":"Some other AS","message":"","query":"2.2.2.2"},{"country":"Some japanese Country"}]`
	if body != expectedBody {
		t.Errorf("\nexpected\n%s\ngot\n%s", expectedBody, body)
	}

	if len(client.Requests) != 1 {
		t.Errorf("expected 1 got %d", len(client.Requests))
	}

	if client.Requests[0] != 3 {
		t.Errorf("expected 3 got %d", client.Requests[0])
	}
}

func TestBatchDefaults(t *testing.T) {
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
	req.SetRequestURI("http://example.com/batch?fields=country,city")
	req.SetBodyString(`
		[
			"1.1.1.1"
		]
	`)
	ctx.Init(&req, nil, nil)

	h.Index(&ctx)

	body := string(ctx.Response.Body())
	expectedBody := `[{"country":"Some Country","city":"Some City"}]`
	if body != expectedBody {
		t.Errorf("\nexpected\n%s\ngot\n%s", expectedBody, body)
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
	req.SetRequestURI("http://example.com/batch?fields=country,city")

	// 150 IPs.
	req.SetBodyString(`
		[
			"1.1.1.1","2.2.2.2","3.3.3.3","4.4.4.4","5.5.5.5","6.6.6.6","7.7.7.7","8.8.8.8","9.9.9.9",
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
			"147.147.147.147","148.148.148.148","149.149.149.149","150.150.150.150"
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
	expectedBody := `[{"country":"Some Country","city":"Some City"},{"country":"Some other Country","city":"Some other City"},{"country":"3.3.3.3en","city":"3.3.3.3en"},{"country":"4.4.4.4en","city":"4.4.4.4en"},{"country":"5.5.5.5en","city":"5.5.5.5en"},{"country":"6.6.6.6en","city":"6.6.6.6en"},{"country":"7.7.7.7en","city":"7.7.7.7en"},{"country":"8.8.8.8en","city":"8.8.8.8en"},{"country":"9.9.9.9en","city":"9.9.9.9en"},{"country":"10.10.10.10en","city":"10.10.10.10en"},{"country":"11.11.11.11en","city":"11.11.11.11en"},{"country":"12.12.12.12en","city":"12.12.12.12en"},{"country":"13.13.13.13en","city":"13.13.13.13en"},{"country":"14.14.14.14en","city":"14.14.14.14en"},{"country":"15.15.15.15en","city":"15.15.15.15en"},{"country":"16.16.16.16en","city":"16.16.16.16en"},{"country":"17.17.17.17en","city":"17.17.17.17en"},{"country":"18.18.18.18en","city":"18.18.18.18en"},{"country":"19.19.19.19en","city":"19.19.19.19en"},{"country":"20.20.20.20en","city":"20.20.20.20en"},{"country":"21.21.21.21en","city":"21.21.21.21en"},{"country":"22.22.22.22en","city":"22.22.22.22en"},{"country":"23.23.23.23en","city":"23.23.23.23en"},{"country":"24.24.24.24en","city":"24.24.24.24en"},{"country":"25.25.25.25en","city":"25.25.25.25en"},{"country":"26.26.26.26en","city":"26.26.26.26en"},{"country":"27.27.27.27en","city":"27.27.27.27en"},{"country":"28.28.28.28en","city":"28.28.28.28en"},{"country":"29.29.29.29en","city":"29.29.29.29en"},{"country":"30.30.30.30en","city":"30.30.30.30en"},{"country":"31.31.31.31en","city":"31.31.31.31en"},{"country":"32.32.32.32en","city":"32.32.32.32en"},{"country":"33.33.33.33en","city":"33.33.33.33en"},{"country":"34.34.34.34en","city":"34.34.34.34en"},{"country":"35.35.35.35en","city":"35.35.35.35en"},{"country":"36.36.36.36en","city":"36.36.36.36en"},{"country":"37.37.37.37en","city":"37.37.37.37en"},{"country":"38.38.38.38en","city":"38.38.38.38en"},{"country":"39.39.39.39en","city":"39.39.39.39en"},{"country":"40.40.40.40en","city":"40.40.40.40en"},{"country":"41.41.41.41en","city":"41.41.41.41en"},{"country":"42.42.42.42en","city":"42.42.42.42en"},{"country":"43.43.43.43en","city":"43.43.43.43en"},{"country":"44.44.44.44en","city":"44.44.44.44en"},{"country":"45.45.45.45en","city":"45.45.45.45en"},{"country":"46.46.46.46en","city":"46.46.46.46en"},{"country":"47.47.47.47en","city":"47.47.47.47en"},{"country":"48.48.48.48en","city":"48.48.48.48en"},{"country":"49.49.49.49en","city":"49.49.49.49en"},{"country":"50.50.50.50en","city":"50.50.50.50en"},{"country":"51.51.51.51en","city":"51.51.51.51en"},{"country":"52.52.52.52en","city":"52.52.52.52en"},{"country":"53.53.53.53en","city":"53.53.53.53en"},{"country":"54.54.54.54en","city":"54.54.54.54en"},{"country":"55.55.55.55en","city":"55.55.55.55en"},{"country":"56.56.56.56en","city":"56.56.56.56en"},{"country":"57.57.57.57en","city":"57.57.57.57en"},{"country":"58.58.58.58en","city":"58.58.58.58en"},{"country":"59.59.59.59en","city":"59.59.59.59en"},{"country":"60.60.60.60en","city":"60.60.60.60en"},{"country":"61.61.61.61en","city":"61.61.61.61en"},{"country":"62.62.62.62en","city":"62.62.62.62en"},{"country":"63.63.63.63en","city":"63.63.63.63en"},{"country":"64.64.64.64en","city":"64.64.64.64en"},{"country":"65.65.65.65en","city":"65.65.65.65en"},{"country":"66.66.66.66en","city":"66.66.66.66en"},{"country":"67.67.67.67en","city":"67.67.67.67en"},{"country":"68.68.68.68en","city":"68.68.68.68en"},{"country":"69.69.69.69en","city":"69.69.69.69en"},{"country":"70.70.70.70en","city":"70.70.70.70en"},{"country":"71.71.71.71en","city":"71.71.71.71en"},{"country":"72.72.72.72en","city":"72.72.72.72en"},{"country":"73.73.73.73en","city":"73.73.73.73en"},{"country":"74.74.74.74en","city":"74.74.74.74en"},{"country":"75.75.75.75en","city":"75.75.75.75en"},{"country":"76.76.76.76en","city":"76.76.76.76en"},{"country":"77.77.77.77en","city":"77.77.77.77en"},{"country":"78.78.78.78en","city":"78.78.78.78en"},{"country":"79.79.79.79en","city":"79.79.79.79en"},{"country":"80.80.80.80en","city":"80.80.80.80en"},{"country":"81.81.81.81en","city":"81.81.81.81en"},{"country":"82.82.82.82en","city":"82.82.82.82en"},{"country":"83.83.83.83en","city":"83.83.83.83en"},{"country":"84.84.84.84en","city":"84.84.84.84en"},{"country":"85.85.85.85en","city":"85.85.85.85en"},{"country":"86.86.86.86en","city":"86.86.86.86en"},{"country":"87.87.87.87en","city":"87.87.87.87en"},{"country":"88.88.88.88en","city":"88.88.88.88en"},{"country":"89.89.89.89en","city":"89.89.89.89en"},{"country":"90.90.90.90en","city":"90.90.90.90en"},{"country":"91.91.91.91en","city":"91.91.91.91en"},{"country":"92.92.92.92en","city":"92.92.92.92en"},{"country":"93.93.93.93en","city":"93.93.93.93en"},{"country":"94.94.94.94en","city":"94.94.94.94en"},{"country":"95.95.95.95en","city":"95.95.95.95en"},{"country":"96.96.96.96en","city":"96.96.96.96en"},{"country":"97.97.97.97en","city":"97.97.97.97en"},{"country":"98.98.98.98en","city":"98.98.98.98en"},{"country":"99.99.99.99en","city":"99.99.99.99en"},{"country":"100.100.100.100en","city":"100.100.100.100en"},{"country":"101.101.101.101en","city":"101.101.101.101en"},{"country":"102.102.102.102en","city":"102.102.102.102en"},{"country":"103.103.103.103en","city":"103.103.103.103en"},{"country":"104.104.104.104en","city":"104.104.104.104en"},{"country":"105.105.105.105en","city":"105.105.105.105en"},{"country":"106.106.106.106en","city":"106.106.106.106en"},{"country":"107.107.107.107en","city":"107.107.107.107en"},{"country":"108.108.108.108en","city":"108.108.108.108en"},{"country":"109.109.109.109en","city":"109.109.109.109en"},{"country":"110.110.110.110en","city":"110.110.110.110en"},{"country":"111.111.111.111en","city":"111.111.111.111en"},{"country":"112.112.112.112en","city":"112.112.112.112en"},{"country":"113.113.113.113en","city":"113.113.113.113en"},{"country":"114.114.114.114en","city":"114.114.114.114en"},{"country":"115.115.115.115en","city":"115.115.115.115en"},{"country":"116.116.116.116en","city":"116.116.116.116en"},{"country":"117.117.117.117en","city":"117.117.117.117en"},{"country":"118.118.118.118en","city":"118.118.118.118en"},{"country":"119.119.119.119en","city":"119.119.119.119en"},{"country":"120.120.120.120en","city":"120.120.120.120en"},{"country":"121.121.121.121en","city":"121.121.121.121en"},{"country":"122.122.122.122en","city":"122.122.122.122en"},{"country":"123.123.123.123en","city":"123.123.123.123en"},{"country":"124.124.124.124en","city":"124.124.124.124en"},{"country":"125.125.125.125en","city":"125.125.125.125en"},{"country":"126.126.126.126en","city":"126.126.126.126en"},{"country":"127.127.127.127en","city":"127.127.127.127en"},{"country":"128.128.128.128en","city":"128.128.128.128en"},{"country":"129.129.129.129en","city":"129.129.129.129en"},{"country":"130.130.130.130en","city":"130.130.130.130en"},{"country":"131.131.131.131en","city":"131.131.131.131en"},{"country":"132.132.132.132en","city":"132.132.132.132en"},{"country":"133.133.133.133en","city":"133.133.133.133en"},{"country":"134.134.134.134en","city":"134.134.134.134en"},{"country":"135.135.135.135en","city":"135.135.135.135en"},{"country":"136.136.136.136en","city":"136.136.136.136en"},{"country":"137.137.137.137en","city":"137.137.137.137en"},{"country":"138.138.138.138en","city":"138.138.138.138en"},{"country":"139.139.139.139en","city":"139.139.139.139en"},{"country":"140.140.140.140en","city":"140.140.140.140en"},{"country":"141.141.141.141en","city":"141.141.141.141en"},{"country":"142.142.142.142en","city":"142.142.142.142en"},{"country":"143.143.143.143en","city":"143.143.143.143en"},{"country":"144.144.144.144en","city":"144.144.144.144en"},{"country":"145.145.145.145en","city":"145.145.145.145en"},{"country":"146.146.146.146en","city":"146.146.146.146en"},{"country":"147.147.147.147en","city":"147.147.147.147en"},{"country":"148.148.148.148en","city":"148.148.148.148en"},{"country":"149.149.149.149en","city":"149.149.149.149en"},{"country":"150.150.150.150en","city":"150.150.150.150en"}]`
	if body != expectedBody {
		t.Errorf("\nexpected\n%s\ngot\n%s", expectedBody, body)
	}

	// The batches could be split multiple ways depending on the timing of the batch processing so don't test this here.
}

func TestBatchPartialFail(t *testing.T) {
	t.Parallel()

	logger := zerolog.New(zerolog.ConsoleWriter{Out: util.ZerologTestWriter{T: t}, NoColor: true})

	cache := cache.New(1000000)
	client := &fetcher.Mock{}
	batches := batch.New(logger.With().Str("part", "batch").Logger(), cache, client)

	h := handlers.Handler{
		Logger:  logger.With().Str("part", "handler").Logger(),
		Batches: batches,
		Client:  client,
	}

	var ctx fasthttp.RequestCtx
	var req fasthttp.Request
	req.SetRequestURI("http://example.com/batch?fields=country,city,status")

	// 150 IPs.
	req.SetBodyString(`
		[
			{
				"query": "asdasd",
				"fields": "status,message"
			},
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
			"147.147.147.147","148.148.148.148"
		]
	`)
	ctx.Init(&req, nil, nil)

	go func() {
		time.Sleep(time.Millisecond * 100)
		batches.Process()
	}()

	h.Index(&ctx)

	contentType := string(ctx.Response.Header.Peek(fasthttp.HeaderContentType))
	expectedContentType := "application/json"
	if contentType != expectedContentType {
		t.Errorf("expected %q got %q", expectedContentType, contentType)
	}

	body := string(ctx.Response.Body())
	// The first IP fails and isn't even added to the batch.
	// The 100 after that fail because they are added to a batch and contain 0.0.0.0en.
	expectedBody := `[{"status":"fail","message":"invalid query"},` + strings.Repeat(`{"status":"fail"},`, 100) + `{"status":"","country":"100.100.100.100en","city":"100.100.100.100en"},{"status":"","country":"101.101.101.101en","city":"101.101.101.101en"},{"status":"","country":"102.102.102.102en","city":"102.102.102.102en"},{"status":"","country":"103.103.103.103en","city":"103.103.103.103en"},{"status":"","country":"104.104.104.104en","city":"104.104.104.104en"},{"status":"","country":"105.105.105.105en","city":"105.105.105.105en"},{"status":"","country":"106.106.106.106en","city":"106.106.106.106en"},{"status":"","country":"107.107.107.107en","city":"107.107.107.107en"},{"status":"","country":"108.108.108.108en","city":"108.108.108.108en"},{"status":"","country":"109.109.109.109en","city":"109.109.109.109en"},{"status":"","country":"110.110.110.110en","city":"110.110.110.110en"},{"status":"","country":"111.111.111.111en","city":"111.111.111.111en"},{"status":"","country":"112.112.112.112en","city":"112.112.112.112en"},{"status":"","country":"113.113.113.113en","city":"113.113.113.113en"},{"status":"","country":"114.114.114.114en","city":"114.114.114.114en"},{"status":"","country":"115.115.115.115en","city":"115.115.115.115en"},{"status":"","country":"116.116.116.116en","city":"116.116.116.116en"},{"status":"","country":"117.117.117.117en","city":"117.117.117.117en"},{"status":"","country":"118.118.118.118en","city":"118.118.118.118en"},{"status":"","country":"119.119.119.119en","city":"119.119.119.119en"},{"status":"","country":"120.120.120.120en","city":"120.120.120.120en"},{"status":"","country":"121.121.121.121en","city":"121.121.121.121en"},{"status":"","country":"122.122.122.122en","city":"122.122.122.122en"},{"status":"","country":"123.123.123.123en","city":"123.123.123.123en"},{"status":"","country":"124.124.124.124en","city":"124.124.124.124en"},{"status":"","country":"125.125.125.125en","city":"125.125.125.125en"},{"status":"","country":"126.126.126.126en","city":"126.126.126.126en"},{"status":"","country":"127.127.127.127en","city":"127.127.127.127en"},{"status":"","country":"128.128.128.128en","city":"128.128.128.128en"},{"status":"","country":"129.129.129.129en","city":"129.129.129.129en"},{"status":"","country":"130.130.130.130en","city":"130.130.130.130en"},{"status":"","country":"131.131.131.131en","city":"131.131.131.131en"},{"status":"","country":"132.132.132.132en","city":"132.132.132.132en"},{"status":"","country":"133.133.133.133en","city":"133.133.133.133en"},{"status":"","country":"134.134.134.134en","city":"134.134.134.134en"},{"status":"","country":"135.135.135.135en","city":"135.135.135.135en"},{"status":"","country":"136.136.136.136en","city":"136.136.136.136en"},{"status":"","country":"137.137.137.137en","city":"137.137.137.137en"},{"status":"","country":"138.138.138.138en","city":"138.138.138.138en"},{"status":"","country":"139.139.139.139en","city":"139.139.139.139en"},{"status":"","country":"140.140.140.140en","city":"140.140.140.140en"},{"status":"","country":"141.141.141.141en","city":"141.141.141.141en"},{"status":"","country":"142.142.142.142en","city":"142.142.142.142en"},{"status":"","country":"143.143.143.143en","city":"143.143.143.143en"},{"status":"","country":"144.144.144.144en","city":"144.144.144.144en"},{"status":"","country":"145.145.145.145en","city":"145.145.145.145en"},{"status":"","country":"146.146.146.146en","city":"146.146.146.146en"},{"status":"","country":"147.147.147.147en","city":"147.147.147.147en"},{"status":"","country":"148.148.148.148en","city":"148.148.148.148en"}]`
	if body != expectedBody {
		t.Errorf("\nexpected\n%s\ngot\n%s", expectedBody, body)
	}
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

					responses = append(responses, fetcher.MockResponseFor(ip+"en").Trim(511))
				}

				var ctx fasthttp.RequestCtx
				var req fasthttp.Request
				req.SetRequestURI("http://example.com/batch?fields=511")
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
