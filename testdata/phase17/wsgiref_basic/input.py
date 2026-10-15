import io
import threading
import urllib.request
from wsgiref import simple_server, util
from wsgiref.headers import Headers
from wsgiref.util import FileWrapper, application_uri, request_uri, setup_testing_defaults, shift_path_info


def app(environ, start_response):
    path = environ["PATH_INFO"]
    if path == "/fail":
        start_response("500 Internal Server Error", [("Content-Type", "text/plain")])
        return [b"oops"]
    body = ("%s %s q=%s" % (environ["REQUEST_METHOD"], path, environ.get("QUERY_STRING", ""))).encode()
    start_response("200 OK", [("Content-Type", "text/plain"), ("Content-Length", str(len(body))), ("X-App", "demo")])
    return [body]


h = Headers([("Content-Type", "text/html")])
h["X-One"] = "1"
h.add_header("Set-Cookie", "a=b", path="/")
h.add_header("Set-Cookie", "c=d")
print(h.get_all("Set-Cookie"), h["x-one"], len(h), "X-One" in h)
print(str(h).splitlines())
del h["X-One"]
print(h.keys())

env = {}
setup_testing_defaults(env)
print(env["REQUEST_METHOD"], env["SERVER_NAME"], env["wsgi.url_scheme"])
print(request_uri(env), application_uri(env))
env["PATH_INFO"] = "/a/b/c"
env["SCRIPT_NAME"] = ""
print(shift_path_info(env), env["SCRIPT_NAME"], env["PATH_INFO"])
print(util.guess_scheme({}), util.is_hop_by_hop("Connection"), util.is_hop_by_hop("X-Foo"))
fw = FileWrapper(io.BytesIO(b"x" * 25), 10)
print([len(c) for c in fw])

class Quiet(simple_server.WSGIRequestHandler):
    def log_message(self, format, *args):
        pass

srv = simple_server.make_server("127.0.0.1", 0, app, handler_class=Quiet)
port = srv.server_address[1]
t = threading.Thread(target=srv.serve_forever, kwargs={"poll_interval": 0.05})
t.start()
r = urllib.request.urlopen("http://127.0.0.1:%d/hello/world?x=1&y=2" % port)
print(r.status, r.read(), r.headers["X-App"])
try:
    urllib.request.urlopen("http://127.0.0.1:%d/fail" % port)
except Exception as e:
    print(type(e).__name__, e)
srv.shutdown()
srv.server_close()
t.join()
