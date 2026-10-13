package link

import "testing"

func TestLinkJoinsTheConcurrencyModules(t *testing.T) {
	mod, err := project(t, map[string]string{"main.py": "import asyncio\nimport threading\nimport queue\nimport signal\nfrom concurrent.futures import ThreadPoolExecutor\n\nasyncio.run(None)\nt = threading.Thread()\nq = queue.Queue()\nex = ThreadPoolExecutor()\nsignal.alarm(0)\n"})
	if err != nil {
		t.Fatal(err)
	}
	got := names(mod)
	for _, want := range []string{"asyncio_pm_run", "threading_pm_Thread", "queue_pm_Queue", "concurrent__futures_pm_ThreadPoolExecutor", "signal_pm_alarm", "contextvars_pm_ContextVar", "_thread_pm_start_new_thread"} {
		if !got[want] {
			t.Errorf("%s is not in the joined module", want)
		}
	}
}

func TestLinkAsyncDefsAreKept(t *testing.T) {
	mod, err := project(t, map[string]string{"main.py": "import asyncio\n\nasync def main():\n    await asyncio.sleep(0)\n\nasyncio.run(main())\n"})
	if err != nil {
		t.Fatal(err)
	}
	got := names(mod)
	if !got["main"] || !got["asyncio_pm_sleep"] {
		t.Errorf("names %v", got)
	}
}
