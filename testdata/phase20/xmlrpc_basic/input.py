import datetime
import threading
import xmlrpc.client
from xmlrpc.server import SimpleXMLRPCServer

body = xmlrpc.client.dumps((1, "two", 3.5, True, [1, 2], {"k": "v"}, None), "echo", allow_none=True)
print(body)
params, method = xmlrpc.client.loads(body)
print(method, params)

dt = xmlrpc.client.DateTime(datetime.datetime(2024, 1, 2, 3, 4, 5))
b = xmlrpc.client.Binary(b"\x00raw")
p, _ = xmlrpc.client.loads(xmlrpc.client.dumps((dt, b)))
print(p[0], p[1].data)

fault = xmlrpc.client.dumps(xmlrpc.client.Fault(4, "Too many"), methodresponse=True)
try:
    xmlrpc.client.loads(fault)
except xmlrpc.client.Fault as f:
    print("Fault", f.faultCode, f.faultString)

server = SimpleXMLRPCServer(("127.0.0.1", 0), logRequests=False, allow_none=True)
server.register_function(lambda a, b: a + b, "add")
server.register_function(pow)
server.register_introspection_functions()
server.register_multicall_functions()
port = server.server_address[1]
t = threading.Thread(target=server.serve_forever)
t.start()
try:
    proxy = xmlrpc.client.ServerProxy(f"http://127.0.0.1:{port}/", allow_none=True)
    print(proxy.add(2, 3), proxy.add("a", "b"), proxy.pow(2, 10))
    print(sorted(proxy.system.listMethods()))
    try:
        proxy.nosuch()
    except xmlrpc.client.Fault as f:
        print("Fault", f.faultCode, f.faultString)
    multi = xmlrpc.client.MultiCall(proxy)
    multi.add(1, 1)
    multi.pow(3, 2)
    print(list(multi()))
finally:
    server.shutdown()
    t.join()
    server.server_close()
