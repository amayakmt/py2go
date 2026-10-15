from pathlib import PurePath, PurePosixPath, Path, PosixPath

p = PurePosixPath("/usr/local//lib/./python3.13/site-packages/")
print(p, repr(p), p.parts, p.drive, repr(p.root), p.anchor, p.name, p.suffix, p.stem)
print(p.parent, p.parents[0], p.parents[2], len(p.parents), list(p.parents)[-1], p.parents[-1], p.parents[:2])
print(PurePath("a/b.tar.gz").suffixes, PurePath("a/b.tar.gz").suffix, PurePath("a/b.tar.gz").stem, PurePath(".bashrc").suffix, PurePath("a/b.").suffix, PurePath("..").name)
print(PurePath(), PurePath("."), PurePath(""), repr(PurePath("")), PurePath("/"), PurePath("//a"), PurePath("///a"), PurePath("a/../b"))
print(PurePath("a") / "b" / PurePath("c"), "x" / PurePath("y"), PurePath("a", "b", "/c", "d"), PurePath("/a").joinpath("b", "c"))
print(PurePath("/a/b").is_absolute(), PurePath("a/b").is_absolute(), PurePath("/a/b").is_relative_to("/a"), PurePath("/a/b").is_relative_to("/c"))
print(PurePath("/a/b/c").relative_to("/a"), PurePath("/a/b/c").relative_to("/a/b/c"), PurePath("/a/b").relative_to("/a/c", walk_up=True))
for bad in [lambda: PurePath("/a").relative_to("/b"), lambda: PurePath("a").relative_to("/b"), lambda: PurePath("a").with_name(""),
            lambda: PurePath("/").with_name("x"), lambda: PurePath("a").with_suffix("x"), lambda: PurePath("a").with_name("b/c"),
            lambda: PurePath(1), lambda: PurePath("/a").relative_to("b", walk_up=True)]:
    try:
        bad()
    except (ValueError, TypeError) as e:
        print(type(e).__name__, e)
print(PurePath("a/b.txt").with_name("c.py"), PurePath("a/b.txt").with_suffix(".md"), PurePath("a/b.txt").with_suffix(""), PurePath("a/b.txt").with_stem("z"))
print(PurePath("a/b.txt").match("*.txt"), PurePath("a/b/c.txt").match("b/*.txt"), PurePath("/a/b/c.txt").match("/a/*.txt"), PurePath("a/b.txt").match("A/*.TXT"), PurePath("a/b.txt").full_match("**/*.txt"), PurePath("a/b/c/d.txt").full_match("a/**/d.txt"))
print(PurePath("a") == PurePath("a/"), PurePath("a") == PurePath("./a"), PurePath("a") == "a", PurePath("a") < PurePath("b"), sorted([PurePath("b"), PurePath("a/c"), PurePath("a")]))
print({PurePath("a"): 1}[PurePath("a")], hash(PurePath("a/b")) == hash(PurePath("a", "b")))
print(PurePath("/a b/c").as_uri(), PurePath("/é").as_uri(), PurePath("a/b").as_posix(), bytes(PurePath("a/b")), str(PurePath("a")), PurePath("a").__fspath__())
print(type(PurePosixPath("a")).__name__, isinstance(PurePosixPath("a"), PurePath))
q = Path("some/dir/file.txt")
print(q, repr(q), q.name, q.parent, q.with_suffix(".bin"), q.is_absolute(), isinstance(q, PurePath), isinstance(q, Path))
print(Path("/tmp/x").absolute(), Path("rel").absolute() == Path.cwd() / "rel", Path("~").expanduser() == Path.home(), Path.home().is_absolute())
print(Path("/a/b/../c").resolve() , Path("/nonexistent/../x").resolve(), Path(".").resolve() == Path.cwd())
print(PosixPath("x/y"), repr(PosixPath("x/y")))
print(PurePath("a/b").parts, PurePath("/").parts, PurePath("/").name, repr(PurePath("/").parent), PurePath("a").parent, PurePath("a").parents[0])
