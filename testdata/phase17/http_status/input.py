import http
from http import HTTPStatus, HTTPMethod

s = HTTPStatus.NOT_FOUND
print(s, repr(s), s.value, s.phrase, s.description, s.name, int(s), s == 404, 404 == s, s < 500, s in (404, 500))
print(HTTPStatus(200), HTTPStatus(200).phrase, HTTPStatus.OK.is_success, s.is_client_error, HTTPStatus.IM_A_TEAPOT.value)
print({HTTPStatus.OK: "ok"}[HTTPStatus.OK], "%d %s" % (HTTPStatus.CREATED, HTTPStatus.CREATED.phrase), f"{HTTPStatus.ACCEPTED}", f"{HTTPStatus.ACCEPTED:>5}")
try:
    HTTPStatus(299)
except ValueError as e:
    print(e)
print(HTTPMethod.GET, repr(HTTPMethod.POST), HTTPMethod.GET == "GET", HTTPMethod.PATCH.description)
print(HTTPStatus.MOVED_PERMANENTLY.value, HTTPStatus.INTERNAL_SERVER_ERROR.phrase, HTTPStatus.SERVICE_UNAVAILABLE.is_server_error)
