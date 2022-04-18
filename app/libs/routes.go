package libs

import (
	"encoding/json"
	"fmt"
	"github.com/ghodss/yaml"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	toml "github.com/pelletier/go-toml"
	"net/http"
	"net/url"
	b64 "encoding/base64"
)

func getSession(c echo.Context) *sessions.Session {
	sess, _ := session.Get("session", c)
	sess.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7,
		HttpOnly: true,
	}

	return sess
}

func MainRoute(c echo.Context) error {
	var tempalateData interface{}
	return c.Render(http.StatusOK, "main", tempalateData)
}

func ConverterRoute(c echo.Context) error {
	sess := getSession(c)

	sess.Save(c.Request(), c.Response())
	return c.Render(http.StatusOK, "converter", map[string]interface{}{
		"from_lang": sess.Values["from_lang"],
		"to_lang":   sess.Values["to_lang"],
		"input":     sess.Values["input"],
		"output":    sess.Values["output"],
	})
}

func ApiConverterRoute(c echo.Context) error {

	sess := getSession(c)

	sess.Values["from_lang"] = c.FormValue("from_lang")
	sess.Values["to_lang"] = c.FormValue("to_lang")

	var input map[string]interface{}
	sess.Values["input"] = c.FormValue("input")
	sess.Values["output"] = ""

	if c.FormValue("from_lang") == "json" {
		err := json.Unmarshal([]byte(c.FormValue("input")), &input)
		if err != nil {
			fmt.Println(err)
		}
	}

	if c.FormValue("from_lang") == "yaml" {
		err := yaml.Unmarshal([]byte(c.FormValue("input")), &input)
		if err != nil {
			fmt.Println(err)
		}
	}

	if c.FormValue("from_lang") == "toml" {
		err := toml.Unmarshal([]byte(c.FormValue("input")), &input)
		if err != nil {
			fmt.Println(err)
		}
	}

	if c.FormValue("to_lang") == "json" {
		stringOutput, err := json.MarshalIndent(input, "", "    ")
		if err != nil {
			fmt.Println(err)
		}

		sess.Values["output"] = fmt.Sprintf("%s", stringOutput)
	}

	if c.FormValue("to_lang") == "yaml" {
		stringOutput, err := yaml.Marshal(input)
		if err != nil {
			fmt.Println(err)
		}

		sess.Values["output"] = fmt.Sprintf("%s", stringOutput)
	}

	if c.FormValue("to_lang") == "toml" {
		stringOutput, err := toml.Marshal(input)
		if err != nil {
			fmt.Println(err)
		}

		sess.Values["output"] = fmt.Sprintf("%s", stringOutput)
	}

	sess.Save(c.Request(), c.Response())

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
		"operation":     sess.Values["operation"],
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
		sess.Values["output"] = fmt.Sprintf("%s", stringOutput)
	}

	if c.FormValue("operation") == "decode" {
		stringOutput, _ := b64.URLEncoding.DecodeString(c.FormValue("input"))
		sess.Values["output"] = fmt.Sprintf("%s", stringOutput)
	}

	sess.Save(c.Request(), c.Response())

	c.Redirect(http.StatusFound, "/base64")
	return nil
}

func HtmlEncodeRoute(c echo.Context) error {
	sess := getSession(c)

	return c.Render(http.StatusOK, "htmlencode", map[string]interface{}{
		"operation":     sess.Values["htmlencode_operation"],
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
		sess.Values["htmlencode_output"] = fmt.Sprintf("%s", stringOutput)
	}

	if c.FormValue("operation") == "decode" {
		stringOutput, _ := url.QueryUnescape(c.FormValue("input"))
		sess.Values["htmlencode_output"] = fmt.Sprintf("%s", stringOutput)
	}

	sess.Save(c.Request(), c.Response())

	c.Redirect(http.StatusFound, "/htmlencode")
	return nil
}
