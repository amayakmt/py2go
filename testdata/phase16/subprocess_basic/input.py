import subprocess
import sys
import os

r = subprocess.run(["echo", "hello", "world"], capture_output=True, text=True)
print(r.returncode, repr(r.stdout), repr(r.stderr), r.args)
print(r)
r = subprocess.run("echo out; echo err >&2; exit 3", shell=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
print(r.returncode, r.stdout, r.stderr)
r = subprocess.run(["cat"], input=b"piped data\n", stdout=subprocess.PIPE)
print(r.stdout, r.returncode)
r = subprocess.run(["cat"], input="text in\n", capture_output=True, text=True)
print(repr(r.stdout))
print(subprocess.check_output(["printf", "a\\nb\\n"]), subprocess.check_output("echo $FOO", shell=True, env={"FOO": "bar", "PATH": os.environ["PATH"]}, text=True))
print(subprocess.call(["true"]), subprocess.call(["false"]), subprocess.call("exit 7", shell=True))
print(subprocess.getoutput("echo hi; echo there"), subprocess.getstatusoutput("echo x; exit 2"))
print(subprocess.check_output(["pwd"], cwd="/tmp", text=True), end="")

try:
    subprocess.check_call(["sh", "-c", "exit 5"])
except subprocess.CalledProcessError as e:
    print(e, e.returncode, e.cmd)
try:
    subprocess.run(["false"], check=True)
except subprocess.CalledProcessError as e:
    print(e.returncode, e.output, e.stdout)
try:
    subprocess.run(["definitely-not-a-program-xyz"])
except FileNotFoundError as e:
    print(type(e).__name__, e.errno, e)
try:
    subprocess.run(["sleep", "5"], timeout=0.2)
except subprocess.TimeoutExpired as e:
    print(e)
try:
    subprocess.run(["sh", "-c", "kill -9 $$"], check=True)
except subprocess.CalledProcessError as e:
    print(e, e.returncode)

p = subprocess.Popen(["cat"], stdin=subprocess.PIPE, stdout=subprocess.PIPE, text=True)
out, err = p.communicate("via popen\n")
print(repr(out), err, p.returncode, p.poll(), isinstance(p.pid, int))
p = subprocess.Popen("echo line1; echo line2", shell=True, stdout=subprocess.PIPE, universal_newlines=True)
print([line for line in p.stdout], p.wait())
p = subprocess.Popen(["sleep", "10"])
print(p.poll())
p.terminate()
print(p.wait(), p.returncode)
with subprocess.Popen(["echo", "ctx"], stdout=subprocess.PIPE) as p:
    print(p.stdout.read())
print(p.returncode)
p = subprocess.Popen(["sh", "-c", "echo to-err >&2"], stderr=subprocess.STDOUT, stdout=subprocess.PIPE)
print(p.communicate())
print(subprocess.run(["sh", "-c", "exit 0"], stdout=subprocess.DEVNULL).returncode)
print(subprocess.list2cmdline(["a b", 'c"d', "e"]))
sys.stdout.flush()
subprocess.run(["echo", "child writes directly"])
print("parent after child")
