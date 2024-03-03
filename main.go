package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/psvmcc/hub/pkg/handlers"
	"github.com/psvmcc/hub/pkg/logging"
	"github.com/psvmcc/hub/pkg/templates"
	"github.com/psvmcc/hub/pkg/types"
	"github.com/psvmcc/hub/pkg/victoriametrics"

	"github.com/VictoriaMetrics/metrics"
	"github.com/labstack/echo/v4"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
)

var (
	version string
	commit  string

	cfg types.ConfigFile
)

func main() {
	app := cli.NewApp()
	app.Name = "HUB"
	app.Usage = "Cashing proxy server"

	app.Version = fmt.Sprintf("%s (%s)", version, commit)
	app.Commands = []*cli.Command{
		{
			Name:    "server",
			Aliases: []string{"s"},
			Usage:   "Starts cashing proxy server",
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:    "verbose",
					Usage:   "Verbose logging",
					EnvVars: []string{"HUB_VERBOSE"},
				},
				&cli.StringFlag{
					Name:    "bind",
					Usage:   "Bind address",
					Value:   "0.0.0.0:6587",
					EnvVars: []string{"HUB_BIND"},
				},
				&cli.StringFlag{
					Name:    "self-exporter-bind",
					Usage:   "Metrics self exporter bind address",
					Value:   "0.0.0.0:6588",
					EnvVars: []string{"HUB_SELF_EXPORTER_BIND"},
				},
				&cli.StringFlag{
					Name:    "config",
					Usage:   "Config path",
					Value:   "config.yaml",
					EnvVars: []string{"HUB_CONFIG"},
				},
			},
			Action: startServer,
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func startServer(c *cli.Context) error {
	metrics.GetOrCreateCounter(fmt.Sprintf("hub_app_version{version=%q,commit=%q}", version, commit)).Inc()
	cfg.Load(c.String("config"))
	logger := logging.Build(c.Bool("verbose"))
	zap.ReplaceGlobals(logger)

	echo.NotFoundHandler = func(c echo.Context) error {
		return c.String(http.StatusNotFound, "404 page not found")
	}
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	e.IPExtractor = echo.ExtractIPFromXFFHeader(
		echo.TrustLoopback(true),
		echo.TrustLinkLocal(false),
		echo.TrustPrivateNet(false),
	)

	httpLogger := zap.S().Named("http")

	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set("cfg", cfg)
			c.Set("logger", httpLogger)
			c.Response().Header().Set("Server", fmt.Sprintf("hub/%s (%s)", version, commit))
			return next(c)
		}
	})

	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			req := c.Request()
			res := c.Response()
			start := time.Now()
			if err = next(c); err != nil {
				c.Error(err)
			}
			stop := time.Now()
			message := fmt.Sprintf(
				"[%s] %s %s requested from %s with status %d in %s [%s]",
				c.Request().Host,
				req.Method,
				req.RequestURI,
				c.RealIP(),
				res.Status,
				stop.Sub(start).String(),
				c.Path(),
			)

			logger := c.Get("logger").(*zap.SugaredLogger)

			if res.Status >= 100 && res.Status <= 399 {
				logger.Named("req").Info(message)
			} else if res.Status >= 400 && res.Status <= 499 {
				logger.Named("req").Warn(message)
			} else {
				logger.Named("req").Error(message)
			}

			return
		}
	})

	e.Renderer = &templates.TemplateRegistry{
		Templates: template.Must(template.New("pypi").Funcs(template.FuncMap{"kindIs": templates.KindIs}).Parse(templates.PypiHTML)),
	}
	e.GET("/*", func(c echo.Context) error {
		return c.String(http.StatusNotFound, "")
	}).Name = "global::default"
	e.GET("/ping", func(c echo.Context) error {
		return c.String(http.StatusOK, "pong")
	}).Name = "global::ping"

	for k := range cfg.Server.PYPI {
		p := e.Group(fmt.Sprintf("/pypi/%s", k))
		p.GET("/simple/:name/", handlers.PypiSimple(k)).Name = fmt.Sprintf("pypi::%s::simple", k)
		p.GET("/packages/:name/:filename", handlers.PypiPackages(k)).Name = fmt.Sprintf("pypi::%s::packages", k)
	}

	for k := range cfg.Server.Static {
		s := e.Group(fmt.Sprintf("/static/%s", k))
		s.GET("/get/*", handlers.Static(k)).Name = fmt.Sprintf("static::%s", k)
	}

	for k, v := range cfg.Server.Galaxy {
		g := e.Group(fmt.Sprintf("/galaxy/%s", k))
		if v.URL != "" && v.Dir != "" {
			log.Fatalf("[GALAXY] Wrong config definition for [%s], please don't use url and dir params together.", k)
		}
		g.Any("", func(c echo.Context) error {
			return c.String(http.StatusOK, "")
		})
		g.GET("/api", func(c echo.Context) error {
			data := types.APIVersions{}
			data.AvailableVersions.V3 = "v3/"
			return c.JSON(http.StatusOK, data)
		}).Name = "galaxy::api"
		if v.URL != "" {
			g.GET("/api/v3/collections/:namespace/:name/", handlers.GalaxyProxyCollection(k)).Name = fmt.Sprintf("galaxy::%s::collection", k)
			g.GET("/api/v3/collections/:namespace/:name/versions/", handlers.GalaxyProxyCollectionVersions(k)).Name = fmt.Sprintf("galaxy::%s::collection::versions", k)
			g.GET("/api/v3/collections/:namespace/:name/versions/:version/", handlers.GalaxyProxyCollectionVersionInfo(k)).Name = fmt.Sprintf("galaxy::%s::collection::version", k)
			g.GET("/get/:namespace/:name/:version", handlers.GalaxyProxyCollectionGet(k)).Name = fmt.Sprintf("galaxy::%s::get", k)
		} else if v.Dir != "" {
			g.GET("/api/v3/collections/:namespace/:name/", handlers.GalaxyLocalCollection(k)).Name = fmt.Sprintf("galaxy::%s::collection", k)
			g.GET("/api/v3/collections/:namespace/:name/versions/", handlers.GalaxyLocalCollectionVersions(k)).Name = fmt.Sprintf("galaxy::%s::collection::versions", k)
			g.GET("/api/v3/collections/:namespace/:name/versions/:version/", handlers.GalaxyLocalCollectionVersionInfo(k)).Name = fmt.Sprintf("galaxy::%s::collection::version", k)
			g.GET("/get/:namespace/:name/:version", handlers.GalaxyLocalCollectionGet(k)).Name = fmt.Sprintf("galaxy::%s::get", k)
		} else {
			log.Fatalf("[GALAXY] Wrong config definition for [%s], please use url or dir param.", k)
		}
	}

	go func() {
		log.Fatal(e.Start(c.String("bind")))
	}()
	return victoriametrics.ListenMetricsServer(c.String("self-exporter-bind"))
}
