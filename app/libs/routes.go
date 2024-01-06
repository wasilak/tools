package libs

import (
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/ghodss/yaml"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	toml "github.com/pelletier/go-toml"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

var tracer = otel.Tracer("tools")

func getSession(c echo.Context) *sessions.Session {
	sess, _ := session.Get("session", c)
	sess.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7,
		HttpOnly: true,
	}

	return sess
}

func handleSessParam(param interface{}) string {
	result, ok := param.(string)
	if !ok {
		// slog.Error("error", "handleSessParam", param)
		return ""
	}

	return result
}

func MainRoute(c echo.Context) error {
	_, span := tracer.Start(c.Request().Context(), "MainRoute")

	var tempalateData interface{}
	span.End()
	return c.Render(http.StatusOK, "main", tempalateData)
}

func ConverterRoute(c echo.Context) error {
	_, span := tracer.Start(c.Request().Context(), "ConverterRoute")
	sess := getSession(c)

	sess.Save(c.Request(), c.Response())

	sessionValues := map[string]string{
		"from_lang": handleSessParam(sess.Values["from_lang"]),
		"to_lang":   handleSessParam(sess.Values["to_lang"]),
		"input":     handleSessParam(sess.Values["input"]),
		"output":    handleSessParam(sess.Values["output"]),
	}

	span.SetAttributes(attribute.String("from_lang", sessionValues["from_lang"]))
	span.SetAttributes(attribute.String("to_lang", sessionValues["to_lang"]))
	span.SetAttributes(attribute.String("input", sessionValues["input"]))
	span.SetAttributes(attribute.String("output", sessionValues["output"]))

	span.End()
	return c.Render(http.StatusOK, "converter", sessionValues)
}

func ApiConverterRoute(c echo.Context) error {
	ctx, span := tracer.Start(c.Request().Context(), "ApiConverterRoute")

	sess := getSession(c)

	sess.Values["from_lang"] = c.FormValue("from_lang")
	sess.Values["to_lang"] = c.FormValue("to_lang")

	var input map[string]interface{}
	sess.Values["input"] = c.FormValue("input")
	sess.Values["output"] = ""

	sessionValues := map[string]string{
		"from_lang": handleSessParam(sess.Values["from_lang"]),
		"to_lang":   handleSessParam(sess.Values["to_lang"]),
		"input":     handleSessParam(sess.Values["input"]),
		"output":    handleSessParam(sess.Values["output"]),
	}

	span.SetAttributes(attribute.String("from_lang", sessionValues["from_lang"]))
	span.SetAttributes(attribute.String("to_lang", sessionValues["to_lang"]))
	span.SetAttributes(attribute.String("input", sessionValues["input"]))
	span.SetAttributes(attribute.String("output", sessionValues["output"]))

	ctx, spanFrom := tracer.Start(ctx, fmt.Sprintf("from_lang_%s", c.FormValue("from_lang")))
	if c.FormValue("from_lang") == "json" {
		err := json.Unmarshal([]byte(c.FormValue("input")), &input)
		if err != nil {
			fmt.Println(err)
			span.SetStatus(codes.Error, fmt.Sprintf("from_lang_%s", c.FormValue("from_lang")))
			span.RecordError(err)
		}
	}

	if c.FormValue("from_lang") == "yaml" {
		err := yaml.Unmarshal([]byte(c.FormValue("input")), &input)
		if err != nil {
			fmt.Println(err)
			span.SetStatus(codes.Error, fmt.Sprintf("from_lang_%s", c.FormValue("from_lang")))
			span.RecordError(err)
		}
	}

	if c.FormValue("from_lang") == "toml" {
		err := toml.Unmarshal([]byte(c.FormValue("input")), &input)
		if err != nil {
			fmt.Println(err)
			span.SetStatus(codes.Error, fmt.Sprintf("from_lang_%s", c.FormValue("from_lang")))
			span.RecordError(err)
		}
	}
	spanFrom.End()

	_, spanTo := tracer.Start(ctx, fmt.Sprintf("to_lang_%s", c.FormValue("to_lang")))
	if c.FormValue("to_lang") == "json" {
		stringOutput, err := json.MarshalIndent(input, "", "    ")
		if err != nil {
			fmt.Println(err)
			span.SetStatus(codes.Error, fmt.Sprintf("to_lang_%s", c.FormValue("to_lang")))
			span.RecordError(err)
		}

		sess.Values["output"] = string(stringOutput)
	}

	if c.FormValue("to_lang") == "yaml" {
		stringOutput, err := yaml.Marshal(input)
		if err != nil {
			fmt.Println(err)
			span.SetStatus(codes.Error, fmt.Sprintf("to_lang_%s", c.FormValue("to_lang")))
			span.RecordError(err)
		}

		sess.Values["output"] = string(stringOutput)
	}

	if c.FormValue("to_lang") == "toml" {
		stringOutput, err := toml.Marshal(input)
		if err != nil {
			fmt.Println(err)
			span.SetStatus(codes.Error, fmt.Sprintf("to_lang_%s", c.FormValue("to_lang")))
			span.RecordError(err)
		}

		sess.Values["output"] = string(stringOutput)
	}
	spanTo.End()

	sess.Save(c.Request(), c.Response())

	span.End()

	c.Redirect(http.StatusFound, "/converter")
	return nil
}

func HealthRoute(c echo.Context) error {
	return c.JSONPretty(http.StatusOK, map[string]interface{}{
		"status": "OK",
	}, "  ")
}

func Base64Route(c echo.Context) error {
	sess := getSession(c)

	return c.Render(http.StatusOK, "base64", map[string]interface{}{
		"operation": sess.Values["operation"],
		"input":     sess.Values["input"],
		"output":    sess.Values["output"],
	})
}

func Base64ApiRoute(c echo.Context) error {
	sess := getSession(c)

	sess.Values["operation"] = c.FormValue("operation")
	sess.Values["input"] = c.FormValue("input")
	sess.Values["output"] = ""

	if c.FormValue("operation") == "encode" {
		stringOutput := b64.URLEncoding.EncodeToString([]byte(c.FormValue("input")))
		sess.Values["output"] = stringOutput
	}

	if c.FormValue("operation") == "decode" {
		stringOutput, _ := b64.URLEncoding.DecodeString(c.FormValue("input"))
		sess.Values["output"] = stringOutput
	}

	sess.Save(c.Request(), c.Response())

	c.Redirect(http.StatusFound, "/base64")
	return nil
}

func HtmlEncodeRoute(c echo.Context) error {
	sess := getSession(c)

	return c.Render(http.StatusOK, "htmlencode", map[string]interface{}{
		"operation": sess.Values["htmlencode_operation"],
		"input":     sess.Values["htmlencode_input"],
		"output":    sess.Values["htmlencode_output"],
	})
}

func HtmlEncodeApiRoute(c echo.Context) error {
	sess := getSession(c)

	sess.Values["htmlencode_operation"] = c.FormValue("operation")
	sess.Values["htmlencode_input"] = c.FormValue("input")
	sess.Values["htmlencode_output"] = ""

	if c.FormValue("operation") == "encode" {
		stringOutput := url.PathEscape(c.FormValue("input"))
		sess.Values["htmlencode_output"] = stringOutput
	}

	if c.FormValue("operation") == "decode" {
		stringOutput, _ := url.QueryUnescape(c.FormValue("input"))
		sess.Values["htmlencode_output"] = stringOutput
	}

	sess.Save(c.Request(), c.Response())

	c.Redirect(http.StatusFound, "/htmlencode")
	return nil
}
