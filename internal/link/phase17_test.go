package link

import "testing"

func TestLinkJoinsTheNetworkModules(t *testing.T) {
	mod, err := project(t, map[string]string{"main.py": "import socket\nimport socketserver\nimport selectors\nimport ipaddress\nimport http.client\nimport http.server\nimport http.cookies\nimport http.cookiejar\nimport urllib.parse\nimport urllib.request\nimport urllib.error\nimport urllib.robotparser\nimport email.message\nimport email.mime.text\nimport smtplib\nimport poplib\nimport ftplib\nimport wsgiref.simple_server\n\nprint(socket.AF_INET, socketserver.TCPServer, selectors.DefaultSelector, ipaddress.ip_address, http.client.HTTPConnection, http.server.HTTPServer, http.cookies.SimpleCookie, http.cookiejar.CookieJar, urllib.parse.urlparse, urllib.request.urlopen, urllib.error.URLError, urllib.robotparser.RobotFileParser, email.message.Message, email.mime.text.MIMEText, smtplib.SMTP, poplib.POP3, ftplib.FTP, wsgiref.simple_server.make_server)\n"})
	if err != nil {
		t.Fatal(err)
	}
	got := names(mod)
	for _, want := range []string{"socket_pm_AF_INET", "socketserver_pm_TCPServer", "ipaddress_pm_ip_address", "http__client_pm_HTTPConnection", "http__server_pm_HTTPServer", "http__cookies_pm_SimpleCookie", "urllib__parse_pm_urlparse", "urllib__request_pm_urlopen", "email__message_pm_Message", "smtplib_pm_SMTP", "poplib_pm_POP3", "ftplib_pm_FTP", "wsgiref__simple_server_pm_make_server"} {
		if !got[want] {
			t.Errorf("%s is not in the joined module", want)
		}
	}
}
