package simplehttp

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

// some good test examples at https://github.com/imroc/req/blob/master/req_test.go

var fixturePath string

func TestMain(m *testing.M) {
	pwd, _ := os.Getwd()
	fixturePath = filepath.Join(pwd, "fixtures")
	os.Exit(m.Run())
}

func TestWrapperMethods(t *testing.T) { //nolint:funlen // subtests for each HTTP method
	// there's gotta be a better way to reference a set of functions and call them sequentially...?
	// cases := []struct {
	// 	method string
	// 	code int
	// }{
	// 	{"get", 200},
	// }

	t.Parallel()
	ts := httptest.NewTLSServer(http.HandlerFunc(handleHTTP))
	defer ts.Close()
	c := New(ts.URL)
	c.Client = ts.Client()

	t.Run("Get", func(t *testing.T) {
		response, err := c.Get("/icanhazdadjoke")
		if err != nil {
			log.Panicln("error:", err)
		}

		got := response
		want := HTTPResponse{
			Body: "",
			Code: 200,
		}

		if !cmp.Equal(want.Code, got.Code) {
			t.Error(cmp.Diff(want.Code, got.Code))
		}
	})

	t.Run("Post", func(t *testing.T) {
		c.Data["key"] = "value"
		response, err := c.Post("/echo")
		if err != nil {
			log.Panicln("error:", err)
		}

		got := response
		wantBody := `{"header":{"Accept-Encoding":["gzip"],` +
			`"Content-Length":["15"],` +
			`"User-Agent":["Go-http-client/1.1"]},` +
			`"body":"{\"key\":\"value\"}"}`
		want := HTTPResponse{
			Body: wantBody,
			Code: 200,
		}

		if !cmp.Equal(want.Code, got.Code) {
			t.Error(cmp.Diff(want.Code, got.Code))
		}
		if !cmp.Equal(want.Body, got.Body) {
			t.Error(cmp.Diff(want.Body, got.Body))
		}
	})

	t.Run("Patch", func(t *testing.T) {
		response, err := c.Patch("/")
		if err != nil {
			log.Panicln("error:", err)
		}

		got := response
		want := HTTPResponse{
			Body: "",
			Code: 200,
		}

		if !cmp.Equal(want.Code, got.Code) {
			t.Error(cmp.Diff(want.Code, got.Code))
		}
	})

	t.Run("Put", func(t *testing.T) {
		response, err := c.Put("/")
		if err != nil {
			log.Panicln("error:", err)
		}

		got := response
		want := HTTPResponse{
			Body: "",
			Code: 200,
		}

		if !cmp.Equal(want.Code, got.Code) {
			t.Error(cmp.Diff(want.Code, got.Code))
		}
	})

	t.Run("Delete", func(t *testing.T) {
		response, err := c.Delete("/")
		if err != nil {
			log.Panicln("error:", err)
		}

		got := response
		want := HTTPResponse{
			Body: "",
			Code: 200,
		}

		if !cmp.Equal(want.Code, got.Code) {
			t.Error(cmp.Diff(want.Code, got.Code))
		}
	})

	t.Run("Head", func(t *testing.T) {
		c.Headers["accept"] = "Application/JSON"
		c.Headers["Content-Type"] = "Application/JSON"

		response, err := c.Head("/header")
		if err != nil {
			log.Panicln("error:", err)
		}

		wantHeaders := make(http.Header)
		wantHeaders.Add("Method", "HEAD")
		wantHeaders.Add("Content-Type", "application/json; charset=utf-8")

		got := response
		want := HTTPResponse{
			Body:    "",
			Code:    200,
			Headers: wantHeaders,
		}

		if !cmp.Equal(want.Code, got.Code) {
			t.Error(cmp.Diff(want.Code, got.Code))
		}

		if !cmp.Equal(want.Headers["Content-Type"], got.Headers["Content-Type"]) {
			t.Error(cmp.Diff(want.Headers, got.Headers))
		}
	})
}

func TestErrorPaths(t *testing.T) { //nolint:funlen // subtests for each error scenario
	t.Parallel()
	ts := httptest.NewTLSServer(http.HandlerFunc(handleHTTP))
	defer ts.Close()
	c := New(ts.URL)
	c.Client = ts.Client()

	t.Run("BadRequest", func(t *testing.T) {
		response, err := c.Get("/bad-request")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if response.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, response.Code)
		}
	})

	t.Run("TooManyRequests", func(t *testing.T) {
		response, err := c.Get("/too-many")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if response.Code != http.StatusTooManyRequests {
			t.Errorf("expected status %d, got %d", http.StatusTooManyRequests, response.Code)
		}
	})

	t.Run("ProtectedWithoutToken", func(t *testing.T) {
		response, err := c.Get("/protected")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if response.Code != http.StatusUnauthorized {
			t.Errorf("expected status %d, got %d", http.StatusUnauthorized, response.Code)
		}
		if response.Body != "bad" {
			t.Errorf("expected body %q, got %q", "bad", response.Body)
		}
	})

	t.Run("ProtectedWithToken", func(t *testing.T) {
		c2 := New(ts.URL)
		c2.Client = ts.Client()
		c2.Headers["Authorization"] = "Bearer goodtoken"

		response, err := c2.Get("/protected")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if response.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, response.Code)
		}
		if response.Body != "good" {
			t.Errorf("expected body %q, got %q", "good", response.Body)
		}
	})

	t.Run("UnlimitedRedirect", func(t *testing.T) {
		_, err := c.Get("/unlimited-redirect")
		if err == nil {
			t.Fatal("expected error for unlimited redirect, got nil")
		}
	})

	t.Run("NilClient", func(t *testing.T) {
		c2 := New(ts.URL)
		c2.Client = nil

		_, err := c2.Get("/")
		if err == nil {
			t.Fatal("expected error for nil client, got nil")
		}
		if !strings.Contains(err.Error(), "http client is nil") {
			t.Errorf("expected nil client error, got: %v", err)
		}
	})
}

func TestQueryParameters(t *testing.T) {
	t.Parallel()
	ts := httptest.NewTLSServer(http.HandlerFunc(handleHTTP))
	defer ts.Close()
	c := New(ts.URL)
	c.Client = ts.Client()

	t.Run("SingleParam", func(t *testing.T) {
		c2 := New(ts.URL)
		c2.Client = ts.Client()
		c2.Params["foo"] = "bar"

		response, err := c2.Get("/query-parameter")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if response.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, response.Code)
		}
		if response.Body != "foo=bar" {
			t.Errorf("expected body %q, got %q", "foo=bar", response.Body)
		}
	})

	t.Run("MultipleParams", func(t *testing.T) {
		c2 := New(ts.URL)
		c2.Client = ts.Client()
		c2.Params["a"] = "1"
		c2.Params["b"] = "2"

		response, err := c2.Get("/query-parameter")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if response.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, response.Code)
		}
		if !strings.Contains(response.Body, "a=1") || !strings.Contains(response.Body, "b=2") {
			t.Errorf("expected query params a=1 and b=2, got %q", response.Body)
		}
	})
}

func TestSetTimeout(t *testing.T) {
	t.Parallel()
	ts := httptest.NewTLSServer(http.HandlerFunc(handleHTTP))
	defer ts.Close()

	t.Run("CustomTimeout", func(t *testing.T) {
		c := New(ts.URL)
		c.Client = ts.Client()
		c.SetTimeout(30 * time.Second)

		if c.Client.Timeout != 30*time.Second {
			t.Errorf("expected timeout %v, got %v", 30*time.Second, c.Client.Timeout)
		}

		response, err := c.Get("/icanhazdadjoke")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if response.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, response.Code)
		}
	})

	t.Run("DefaultTimeout", func(t *testing.T) {
		c := New(ts.URL)
		if c.Client.Timeout != 10*time.Second {
			t.Errorf("expected default timeout %v, got %v", 10*time.Second, c.Client.Timeout)
		}
	})
}

func handleHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Method", r.Method)
	switch r.Method {
	case http.MethodGet:
		handleGet(w, r)
	case http.MethodPost:
		handlePost(w, r)
	case http.MethodHead:
		handleHead(w, r)
	}
}

func handleGet(w http.ResponseWriter, r *http.Request) { //nolint:funlen // test handler with many routes
	switch r.URL.Path {
	case "/icanhazdadjoke":
		f, err := os.ReadFile(fmt.Sprintf("%s/icanhazdadjoke.json", fixturePath))
		if err != nil {
			log.Panicf("Error. %+v", err)
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_, _ = w.Write(f)
	case "/bad-request":
		w.WriteHeader(http.StatusBadRequest)
	case "/too-many":
		w.WriteHeader(http.StatusTooManyRequests)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_, _ = w.Write([]byte(`{"errMsg":"too many requests"}`))
	case "/chunked":
		w.Header().Add("Trailer", "Expires")
		_, _ = w.Write([]byte(`This is a chunked body`))
	case "/host-header":
		_, _ = w.Write([]byte(r.Host))
	case "/json":
		_ = r.ParseForm()
		if r.FormValue("type") != "no" {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		if r.FormValue("error") == "yes" {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"message": "not allowed"}`))
		} else {
			_, _ = w.Write([]byte(`{"name": "roc"}`))
		}
	case "/unlimited-redirect":
		w.Header().Set("Location", "/unlimited-redirect")
		w.WriteHeader(http.StatusMovedPermanently)
	case "/redirect-to-other":
		w.Header().Set("Location", "http://dummy.local/test")
		w.WriteHeader(http.StatusMovedPermanently)
	case "/pragma":
		w.Header().Add("Pragma", "no-cache")
	case "/payload":
		b, _ := io.ReadAll(r.Body)
		_, _ = w.Write(b)
	case "/header":
		b, _ := json.Marshal(r.Header)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_, _ = w.Write(b)
	case "/content-type":
		_, _ = w.Write([]byte(r.Header.Get("Content-Type")))
	case "/query-parameter":
		_, _ = w.Write([]byte(r.URL.RawQuery))
	case "/download":
		size := 100 * 1024 * 1024
		w.Header().Set("Content-Length", strconv.Itoa(size))
		buf := make([]byte, 1024)
		for i := 0; i < 1024; i++ {
			buf[i] = 'h'
		}
		for i := 0; i < size; {
			wbuf := buf
			if size-i < 1024 {
				wbuf = buf[:size-i]
			}
			n, err := w.Write(wbuf)
			if err != nil {
				break
			}
			i += n
		}
	case "/protected":
		auth := r.Header.Get("Authorization")
		if auth == "Bearer goodtoken" {
			_, _ = w.Write([]byte("good"))
		} else {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`bad`))
		}
	}
}

// Echo is used in "/echo" API.
type Echo struct {
	Header http.Header `json:"header" xml:"header"`
	Body   string      `json:"body" xml:"body"`
}

func handlePost(w http.ResponseWriter, r *http.Request) { //nolint:funlen // test handler with many routes
	switch r.URL.Path {
	case "/":
		_, _ = io.Copy(io.Discard, r.Body)
		_, _ = w.Write([]byte("TestPost: text response"))
	case "/raw-upload":
		_, _ = io.Copy(io.Discard, r.Body)
	case "/file-text":
		_ = r.ParseMultipartForm(10e6)
		files := r.MultipartForm.File["file"]
		file, _ := files[0].Open()
		b, _ := io.ReadAll(file)
		_ = r.ParseForm()
		if a := r.FormValue("attempt"); a != "" && a != "2" {
			w.WriteHeader(http.StatusInternalServerError)
		}
		_, _ = w.Write(b)
	case "/form":
		_ = r.ParseForm()
		ret, _ := json.Marshal(&r.Form)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_, _ = w.Write(ret)
	case "/multipart":
		_ = r.ParseMultipartForm(10e6)
		m := make(map[string]any)
		m["values"] = r.MultipartForm.Value
		m["files"] = r.MultipartForm.File
		ret, _ := json.Marshal(&m)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_, _ = w.Write(ret)
	case "/redirect":
		_, _ = io.Copy(io.Discard, r.Body)
		w.Header().Set("Location", "/")
		w.WriteHeader(http.StatusMovedPermanently)
	case "/content-type":
		_, _ = io.Copy(io.Discard, r.Body)
		_, _ = w.Write([]byte(r.Header.Get("Content-Type")))
	case "/echo":
		b, _ := io.ReadAll(r.Body)
		e := Echo{
			Header: r.Header,
			Body:   string(b),
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		result, _ := json.Marshal(&e)
		_, _ = w.Write(result)
	}
}

func handleHead(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/header" {
		b, _ := json.Marshal(r.Header)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_, _ = w.Write(b)
	}
}

// online-only test; might use this eventually if I can mock out the correct return values
// func TestWeather(t *testing.T) {
// 	type WeatherResponse struct {
// 		CurrentCondition []struct {
// 			FeelsLikeC string `json:"FeelsLikeC,omitempty"`
// 			FeelsLikeF string `json:"FeelsLikeF,omitempty"`
// 			Cloudcover string `json:"cloudcover,omitempty"`
// 			Humidity   string `json:"humidity,omitempty"`
// 			LangFr     []struct {
// 				Value string `json:"value,omitempty"`
// 			} `json:"lang_fr,omitempty"`
// 			LocalObsDateTime string `json:"localObsDateTime,omitempty"`
// 			ObservationTime  string `json:"observation_time,omitempty"`
// 			PrecipInches     string `json:"precipInches,omitempty"`
// 			PrecipMM         string `json:"precipMM,omitempty"`
// 			Pressure         string `json:"pressure,omitempty"`
// 			PressureInches   string `json:"pressureInches,omitempty"`
// 			TempC            string `json:"temp_C,omitempty"`
// 			TempF            string `json:"temp_F,omitempty"`
// 			UvIndex          string `json:"uvIndex,omitempty"`
// 			Visibility       string `json:"visibility,omitempty"`
// 			VisibilityMiles  string `json:"visibilityMiles,omitempty"`
// 			WeatherCode      string `json:"weatherCode,omitempty"`
// 			WeatherDesc      []struct {
// 				Value string `json:"value,omitempty"`
// 			} `json:"weatherDesc,omitempty"`
// 			WeatherIconURL []struct {
// 				Value string `json:"value,omitempty"`
// 			} `json:"weatherIconUrl,omitempty"`
// 			Winddir16Point string `json:"winddir16Point,omitempty"`
// 			WinddirDegree  string `json:"winddirDegree,omitempty"`
// 			WindspeedKmph  string `json:"windspeedKmph,omitempty"`
// 			WindspeedMiles string `json:"windspeedMiles,omitempty"`
// 		} `json:"current_condition,omitempty"`
// 		NearestArea []struct {
// 			AreaName []struct {
// 				Value string `json:"value,omitempty"`
// 			} `json:"areaName,omitempty"`
// 			Country []struct {
// 				Value string `json:"value,omitempty"`
// 			} `json:"country,omitempty"`
// 			Latitude   string `json:"latitude,omitempty"`
// 			Longitude  string `json:"longitude,omitempty"`
// 			Population string `json:"population,omitempty"`
// 			Region     []struct {
// 				Value string `json:"value,omitempty"`
// 			} `json:"region,omitempty"`
// 			WeatherURL []struct {
// 				Value string `json:"value,omitempty"`
// 			} `json:"weatherUrl,omitempty"`
// 		} `json:"nearest_area,omitempty"`
// 		Request []struct {
// 			Query string `json:"query,omitempty"`
// 			Type  string `json:"type,omitempty"`
// 		} `json:"request,omitempty"`
// 		Weather []struct {
// 			Astronomy []struct {
// 				MoonIllumination string `json:"moon_illumination,omitempty"`
// 				MoonPhase        string `json:"moon_phase,omitempty"`
// 				Moonrise         string `json:"moonrise,omitempty"`
// 				Moonset          string `json:"moonset,omitempty"`
// 				Sunrise          string `json:"sunrise,omitempty"`
// 				Sunset           string `json:"sunset,omitempty"`
// 			} `json:"astronomy,omitempty"`
// 			AvgtempC string `json:"avgtempC,omitempty"`
// 			AvgtempF string `json:"avgtempF,omitempty"`
// 			Date     string `json:"date,omitempty"`
// 			Hourly   []struct {
// 				DewPointC        string `json:"DewPointC,omitempty"`
// 				DewPointF        string `json:"DewPointF,omitempty"`
// 				FeelsLikeC       string `json:"FeelsLikeC,omitempty"`
// 				FeelsLikeF       string `json:"FeelsLikeF,omitempty"`
// 				HeatIndexC       string `json:"HeatIndexC,omitempty"`
// 				HeatIndexF       string `json:"HeatIndexF,omitempty"`
// 				WindChillC       string `json:"WindChillC,omitempty"`
// 				WindChillF       string `json:"WindChillF,omitempty"`
// 				WindGustKmph     string `json:"WindGustKmph,omitempty"`
// 				WindGustMiles    string `json:"WindGustMiles,omitempty"`
// 				Chanceoffog      string `json:"chanceoffog,omitempty"`
// 				Chanceoffrost    string `json:"chanceoffrost,omitempty"`
// 				Chanceofhightemp string `json:"chanceofhightemp,omitempty"`
// 				Chanceofovercast string `json:"chanceofovercast,omitempty"`
// 				Chanceofrain     string `json:"chanceofrain,omitempty"`
// 				Chanceofremdry   string `json:"chanceofremdry,omitempty"`
// 				Chanceofsnow     string `json:"chanceofsnow,omitempty"`
// 				Chanceofsunshine string `json:"chanceofsunshine,omitempty"`
// 				Chanceofthunder  string `json:"chanceofthunder,omitempty"`
// 				Chanceofwindy    string `json:"chanceofwindy,omitempty"`
// 				Cloudcover       string `json:"cloudcover,omitempty"`
// 				Humidity         string `json:"humidity,omitempty"`
// 				LangFr           []struct {
// 					Value string `json:"value,omitempty"`
// 				} `json:"lang_fr,omitempty"`
// 				PrecipInches    string `json:"precipInches,omitempty"`
// 				PrecipMM        string `json:"precipMM,omitempty"`
// 				Pressure        string `json:"pressure,omitempty"`
// 				PressureInches  string `json:"pressureInches,omitempty"`
// 				TempC           string `json:"tempC,omitempty"`
// 				TempF           string `json:"tempF,omitempty"`
// 				Time            string `json:"time,omitempty"`
// 				UvIndex         string `json:"uvIndex,omitempty"`
// 				Visibility      string `json:"visibility,omitempty"`
// 				VisibilityMiles string `json:"visibilityMiles,omitempty"`
// 				WeatherCode     string `json:"weatherCode,omitempty"`
// 				WeatherDesc     []struct {
// 					Value string `json:"value,omitempty"`
// 				} `json:"weatherDesc,omitempty"`
// 				WeatherIconURL []struct {
// 					Value string `json:"value,omitempty"`
// 				} `json:"weatherIconUrl,omitempty"`
// 				Winddir16Point string `json:"winddir16Point,omitempty"`
// 				WinddirDegree  string `json:"winddirDegree,omitempty"`
// 				WindspeedKmph  string `json:"windspeedKmph,omitempty"`
// 				WindspeedMiles string `json:"windspeedMiles,omitempty"`
// 			} `json:"hourly,omitempty"`
// 			MaxtempC    string `json:"maxtempC,omitempty"`
// 			MaxtempF    string `json:"maxtempF,omitempty"`
// 			MintempC    string `json:"mintempC,omitempty"`
// 			MintempF    string `json:"mintempF,omitempty"`
// 			SunHour     string `json:"sunHour,omitempty"`
// 			TotalSnowCm string `json:"totalSnow_cm,omitempty"`
// 			UvIndex     string `json:"uvIndex,omitempty"`
// 		} `json:"weather,omitempty"`
// 	}

// 	c := NewClient()
// 	// Get weather in French JSON
// 	c.BaseURL = "http://wttr.in"
// 	c.Headers["Accept-Language"] = "fr"
// 	c.Params["format"] = "j1"
// 	response, err := c.Get("/55328")
// 	if err != nil {
// 		log.Panicln("error:", err)
// 	}
// 	got := response
// 	want := HTTPResponse{
// 		Body: "",
// 		Code: 200,
// 	}
// 	if !cmp.Equal(want.Code, got.Code) {
// 		t.Error(cmp.Diff(want, got))
// 	}

// 	var gotbody WeatherResponse
// 	unmarshall_error := json.Unmarshal([]byte(got.Body), &gotbody)
// 	if unmarshall_error != nil {
// 		log.Panicln("error:", err)
// 	}
// 	log.Println("type:", gotbody.Request)
// }
