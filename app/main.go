package main

import (
	"embed"
	"flag"
	"html/template"
	"io"
	"io/fs"
	"strings"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/prometheus"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/wasilak/tools/libs"
)

var err error

//go:embed views
var views embed.FS

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func getEmbededViews() fs.FS {
	fsys, err := fs.Sub(views, "views")
	if err != nil {
		panic(err)
	}

	return fsys
}

func main() {
	// using standard library "flag" package
	flag.Bool("verbose", false, "verbose")
	flag.String("listen", "127.0.0.1:3000", "listen address")

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)

	viper.SetEnvPrefix("tools")
	viper.AutomaticEnv()

	log.Debug(viper.AllSettings())

	if viper.GetBool("debug") {
		log.SetLevel(log.DEBUG)
	}

	e := echo.New()

	e.Use(session.Middleware(sessions.NewCookieStore([]byte("secret"))))

	e.Use(middleware.GzipWithConfig(middleware.GzipConfig{
		Skipper: func(c echo.Context) bool {
			return strings.Contains(c.Path(), "metrics")
		},
	}))

	e.HideBanner = true

	if viper.GetBool("verbose") {
		e.Logger.SetLevel(log.DEBUG)
	}

	e.Debug = viper.GetBool("debug")

	t := &Template{
		templates: template.Must(template.ParseFS(getEmbededViews(), "*.html")),
	}

	e.Renderer = t

	e.Use(middleware.Logger())

	// Enable metrics middleware
	p := prometheus.NewPrometheus("echo", nil)
	p.Use(e)

	e.GET("/", libs.MainRoute)
	e.GET("/health", libs.HealthRoute)
	e.GET("/converter", libs.ConverterRoute)
	e.POST("/api/converter", libs.ApiConverterRoute)
	e.GET("/base64", libs.Base64Route)
	e.POST("/api/base64", libs.Base64ApiRoute)
	e.GET("/htmlencode", libs.HtmlEncodeRoute)
	e.POST("/api/htmlencode", libs.HtmlEncodeApiRoute)

	e.Logger.Fatal(e.Start(viper.GetString("listen")))
}
