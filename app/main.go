package main

import (
	"context"
	"embed"
	"flag"
	"html/template"
	"io"
	"io/fs"
	"log/slog"
	"os"
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

	slogecho "github.com/samber/slog-echo"
	"github.com/wasilak/loggergo"
	loggergoLib "github.com/wasilak/loggergo/lib"
	loggergoTypes "github.com/wasilak/loggergo/lib/types"

	otelgotracer "github.com/wasilak/otelgo/tracing"
	"github.com/wasilak/profilego"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
)

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
	flag.String("listen", "127.0.0.1:3000", "listen address")
	flag.String("log.level", "info", "Log level")
	flag.String("log.format", "json", "Log format")
	flag.Bool("otel.enabled", false, "OpenTelemetry traces enabled")
	flag.Bool("profiling.enabled", false, "Profiling enabled")
	flag.String("profiling.ServerAddress", "", "Profiler ServerAddress")

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)

	viper.SetEnvPrefix("tools")
	viper.AutomaticEnv()

	slog.Debug("AllSettings", "value", viper.AllSettings())

	ctx := context.Background()

	if viper.GetBool("otel.enabled") {
		otelGoTracingConfig := otelgotracer.Config{
			HostMetricsEnabled:    true,
			RuntimeMetricsEnabled: true,
		}
		ctx, _, err := otelgotracer.Init(ctx, otelGoTracingConfig)
		if err != nil {
			slog.ErrorContext(ctx, err.Error())
			os.Exit(1)
		}
	}

	loggerConfig := loggergoTypes.Config{
		Level:  loggergoLib.LogLevelFromString(viper.GetString("log.level")),
		Format: loggergoLib.LogFormatFromString(viper.GetString("log.format")),
	}

	if viper.GetBool("otel.enabled") {
		loggerConfig.OtelLoggerName = "github.com/wasilak/tools"
		loggerConfig.OtelServiceName = libs.GetAppName()
		loggerConfig.OtelTracingEnabled = true
	}

	ctx, _, err := loggergo.Init(ctx, loggerConfig)
	if err != nil {
		slog.ErrorContext(ctx, err.Error())
		os.Exit(1)
	}

	if viper.GetBool("profiling.enabled") {
		ProfileGoConfig := profilego.Config{
			ApplicationName: libs.GetAppName(),
			ServerAddress:   viper.GetString("profiling.ServerAddress"),
			Tags: map[string]string{
				"hostname": os.Getenv("HOSTNAME"),
				"test":     "test_value",
			},
		}

		err := profilego.Init(ProfileGoConfig)
		if err != nil {
			slog.ErrorContext(ctx, err.Error())
			os.Exit(1)
		}
	}

	e := echo.New()

	if viper.GetBool("otel.enabled") {
		e.Use(otelecho.Middleware(os.Getenv("OTEL_SERVICE_NAME")))
	}

	e.Use(session.Middleware(sessions.NewCookieStore([]byte("secret"))))

	e.Use(middleware.GzipWithConfig(middleware.GzipConfig{
		Skipper: func(c echo.Context) bool {
			return strings.Contains(c.Path(), "metrics")
		},
	}))

	e.HideBanner = true

	if strings.EqualFold(viper.GetString("log.level"), "debug") {
		e.Logger.SetLevel(log.DEBUG)
		e.Debug = true
	}

	t := &Template{
		templates: template.Must(template.ParseFS(getEmbededViews(), "*.html")),
	}

	e.Renderer = t

	e.Use(slogecho.New(slog.Default()))
	e.Use(middleware.Recover())

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
