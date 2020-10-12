import socket
import logging
import os
import json
from ruamel.yaml import YAML
from ruamel.yaml.compat import StringIO
import toml
from fastapi import FastAPI, Form
from fastapi.templating import Jinja2Templates
from fastapi.responses import HTMLResponse
from starlette.requests import Request
from starlette.middleware.sessions import SessionMiddleware
from starlette.responses import RedirectResponse

app = FastAPI()

logger = logging.getLogger("api")

app.add_middleware(SessionMiddleware, secret_key=os.environ.get("SECRET_KEY", default=""),
                   session_cookie=os.environ.get(
                       "SESSION_COOKIE_NAME", default="session"))

templates = Jinja2Templates(directory="templates")

yaml = YAML()


@app.get('/health')
async def health():
    return {
        "status": "OK"
    }


@app.get("/", response_class=HTMLResponse)
async def root(request: Request):
    return templates.TemplateResponse("root.html", {"request": request})


def decode_json(input):
    return json.loads(input)


def encode_json(input):
    return json.dumps(input, sort_keys=True, indent=4, separators=(',', ': '))


def decode_toml(input):
    return toml.loads(input)


def encode_toml(input):
    return toml.dumps(input)


def decode_yaml(input):
    return yaml.load(input)


def encode_yaml(input):
    string_stream = StringIO()
    yaml.dump(input, string_stream)
    output_str = string_stream.getvalue()
    string_stream.close()
    return output_str


@app.post("/api/converter")
async def api_converter(request: Request, input: str = Form(...), from_lang: str = Form(...), to_lang: str = Form(...)):
    request.session["converter_error"] = ""
    request.session["converter_from_lang"] = from_lang
    request.session["converter_to_lang"] = to_lang

    try:
        request.session["converter_input"] = input.strip()

        input_parsed = ""
        output_parsed = ""

        if from_lang == "json":
            input_parsed = decode_json(input)

        if from_lang == "toml":
            input_parsed = decode_toml(input)

        if from_lang == "yaml":
            input_parsed = decode_yaml(input)

        if to_lang == "json":
            output_parsed = encode_json(input_parsed)

        if to_lang == "yaml":
            output_parsed = encode_yaml(input_parsed)

        if to_lang == "toml":
            output_parsed = encode_toml(input_parsed)

        request.session["converter_output"] = output_parsed.strip()
    except Exception as e:
        request.session["converter_input"] = input.strip()
        request.session["converter_output"] = ""
        request.session["converter_error"] = str(e)

    url = app.url_path_for("converter")
    return RedirectResponse(url=url, status_code=303)  # HTTP_303_SEE_OTHER


@app.get("/converter", response_class=HTMLResponse)
async def converter(request: Request):
    input = request.session["converter_input"] if "converter_input" in request.session else ""

    request.session["converter_input"] = input

    output = request.session["converter_output"] if "converter_output" in request.session else ""
    error = request.session["converter_error"] if "converter_error" in request.session else ""
    from_lang = request.session["converter_from_lang"] if "converter_from_lang" in request.session else ""
    to_lang = request.session["converter_to_lang"] if "converter_to_lang" in request.session else ""
    return templates.TemplateResponse("converter.html", {"request": request, "input": input, "output": output, "from_lang": from_lang, "to_lang": to_lang, "error": error})
