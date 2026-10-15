# Benchmarks

Recorded with `scripts/bench.sh` on 2026-10-16: Go 1.27.1, linux/amd64 (WSL2),
`CGO_ENABLED=0`, CPython 3.13.15 through `uv run`. The CPython column includes
about 0.1 s of `uv` start-up, so small programs are not a fair comparison with it.
The programs are in `bench/`; the transpile time includes the Go build.

| program | what it does | transpile+build (s) | binary (KiB) | py2go run (s) | CPython run (s) |
|---|---|---|---|---|---|
| hello | `print("hello, world")` | 0.51 | 9754 | 0.010 | 0.112 |
| fib | recursive `fib(32)`, annotated | 0.52 | 9759 | 0.034 | 0.288 |
| classes | 2 000 000 method calls creating objects, annotated | 0.55 | 9771 | 0.053 | 0.402 |
| words | dict, string and sort work on 200 000 words, no annotations | 0.68 | 9842 | 0.359 | 0.184 |
| loops | 9 000 000 iterations of integer arithmetic, no annotations | 0.46 | 9765 | 0.859 | 0.616 |
| nbody | 3 000 000 float steps, no annotations | 0.80 | 9781 | 1.896 | 0.475 |

What the numbers say:

- Start-up of a binary is 10 ms (hello world); a binary is about 9.8 MiB.
- Annotated code compiles to typed Go and is 5 to 8 times faster than CPython.
- **Code without annotations is slower than CPython on numeric loops** (1.4x on
  `loops`, 4x on `nbody`) and slower on string and dict work (2x on `words`):
  every operation goes through the dynamic runtime (`DynBinary`, boxed `any`
  values), and the interpreter it competes with has had thirty years of
  tuning. This is the main performance gap; closing it means widening type
  inference so that fewer variables fall back to `object`.
- Transpile plus build is about half a second for a program that imports
  little; programs that import large standard library modules (`argparse`,
  `asyncio`, `unittest`) take several seconds, since the Go compiler builds all
  the joined modules (there is no module cache across programs yet).

The pyperformance suite is not run: it needs `pyperf` and a set of benchmarks
that use `eval`, `exec` and third-party packages. There is no committed baseline
nor CI job that fails on regression; `scripts/bench.sh` prints a table that can
be compared by hand.
