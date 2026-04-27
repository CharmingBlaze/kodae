#!/usr/bin/env python3
"""
Generate Clio structs + extern declarations from raylib.h (tested with raylib 6.x).

  python scripts/gen_raylib_clio.py path/to/raylib.h -o include/raylib/raylib.clio

- Extracts typedef struct { ... } Name; bodies that contain no pointers or C arrays.
- Emits Clio struct fields using f32 / i32 / u32 / u8 and nested struct types.
- Emits RLAPI functions whose signatures only use those structs + primitives + ptr[byte].
- Skips: function pointers, varargs, [ in params, unknown C types, structs with * or [ in body.
"""
from __future__ import annotations

import argparse
import re
import sys
from collections import defaultdict, deque
from pathlib import Path

# C primitive / known -> Clio type for extern params and struct fields
def map_field_type(t: str) -> str | None:
    t = " ".join(t.split())
    t = re.sub(r"^const\s+", "", t).strip()
    if t.endswith("*"):
        return None
    if t in ("float",):
        return "f32"
    if t in ("double",):
        return "float"
    if t in ("int", "signed int"):
        return "i32"
    if t in ("unsigned int", "unsigned"):
        return "u32"
    if t in ("unsigned char",):
        return "u8"
    if t in ("char", "signed char"):
        return "i32"
    if t in ("bool",):
        return "bool"
    if t in ("void",):
        return None
    return None  # struct name handled by caller


def clio_field_line(c_type: str, fname: str, struct_names: set[str]) -> str | None:
    c_type = " ".join(c_type.split())
    c_type = re.sub(r"^const\s+", "", c_type).strip()
    base = map_field_type(c_type)
    if base is not None:
        return f"  {fname}: {base}"
    tok = c_type.replace("  ", " ").strip()
    if tok in struct_names:
        return f"  {fname}: {tok}"
    return None


def parse_structs(text: str) -> dict[str, str]:
    """name -> Clio struct body (fields only, newline-separated)."""
    pat = re.compile(
        r"typedef\s+struct\s+(\w+)\s*\{([^}]+)\}\s*\1\s*;",
        re.MULTILINE | re.DOTALL,
    )
    raw: list[tuple[str, str]] = []
    for m in pat.finditer(text):
        raw.append((m.group(1), m.group(2)))

    all_names = {n for n, _ in raw}
    out: dict[str, str] = {}
    for name, body in raw:
        if "*" in body or "[" in body:
            continue
        lines = []
        for part in body.split(";"):
            part = part.strip()
            if not part:
                continue
            part = re.sub(r"//.*", "", part).strip()
            if not part:
                continue
            cm = re.search(r"^(.+?)\s+(\w+)\s*$", part.replace("\n", " "))
            if not cm:
                continue
            ct, fn = cm.group(1).strip(), cm.group(2).strip()
            lines.append((ct, fn))
        if not lines:
            continue
        field_strs: list[str] = []
        ok = True
        for ct, fn in lines:
            line = clio_field_line(ct, fn, all_names)
            if line is None:
                ok = False
                break
            field_strs.append(line)
        if ok and field_strs:
            out[name] = "\n".join(field_strs)

    for m in re.finditer(r"typedef\s+(\w+)\s+(\w+)\s*;", text):
        a, b = m.group(1), m.group(2)
        if a in out and b not in out:
            out[b] = out[a]

    # Drop structs that reference another struct we did not emit (e.g. Image has void*).
    prim = {"f32", "float", "i32", "u32", "u8", "bool"}
    while True:
        removed = False
        for name, body in list(out.items()):
            for line in body.splitlines():
                m = re.match(r"^\s+\w+:\s+(\w+)\s*$", line)
                if not m:
                    continue
                t = m.group(1)
                if t not in prim and t not in out:
                    del out[name]
                    removed = True
                    break
        if not removed:
            break
    return out


def topo_structs(structs: dict[str, str]) -> list[str]:
    """Order struct names so dependencies appear first (simple DFS)."""
    deps: dict[str, set[str]] = defaultdict(set)
    for name, body in structs.items():
        for other in structs:
            if other != name and re.search(rf"\b{re.escape(other)}\b", body):
                deps[name].add(other)

    order: list[str] = []
    seen: set[str] = set()

    def visit(n: str) -> None:
        if n in seen:
            return
        for d in deps.get(n, ()):
            visit(d)
        seen.add(n)
        order.append(n)

    for n in structs:
        visit(n)
    return order


def map_param_or_ret(t: str, struct_names: set[str]) -> str | None:
    t = " ".join(t.split())
    t = re.sub(r"^const\s+", "", t).strip()
    if t == "void":
        return "void"
    if "*" in t:
        comp = t.replace(" ", "").lower()
        if "char*" in comp or comp == "void*" or comp.startswith("void*"):
            return "ptr[byte]"
        return None
    b = map_field_type(t)
    if b is not None:
        return b
    if re.fullmatch(r"[A-Za-z_]\w*", t):
        if t in struct_names:
            return t
        return None
    return None


def parse_rlapi(text: str, struct_names: set[str]) -> tuple[list[str], int]:
    decls: list[str] = []
    skipped = 0
    for line in text.splitlines():
        s = line.strip()
        if not s.startswith("RLAPI"):
            continue
        s = s.split("//")[0].strip()
        if "(*" in s or not s.endswith(");"):
            skipped += 1
            continue
        if "(" not in s or ")" not in s:
            skipped += 1
            continue
        rest = s[5:].strip()
        try:
            p0 = rest.index("(")
            p1 = rest.rindex(")")
        except ValueError:
            skipped += 1
            continue
        head, arg_str = rest[:p0], rest[p0 + 1 : p1]
        m = re.search(r"([a-zA-Z_][a-zA-Z0-9_]*)\s*$", head)
        if not m:
            skipped += 1
            continue
        fn = m.group(1)
        ret_c = head[: m.start()].strip()
        if "..." in arg_str or "[" in arg_str:
            skipped += 1
            continue
        if "(*" in arg_str:
            skipped += 1
            continue
        rclio = map_param_or_ret(ret_c, struct_names)
        if rclio is None:
            skipped += 1
            continue
        params: list[str] = []
        ok = True
        if arg_str.strip() and arg_str.strip() != "void":
            for raw in split_args(arg_str):
                if not raw:
                    continue
                pm = re.search(r"([a-zA-Z_][a-zA-Z0-9_]*)\s*$", raw)
                if not pm:
                    ok = False
                    break
                pname = pm.group(1)
                tstr = raw[: pm.start()].strip()
                ct = map_param_or_ret(tstr, struct_names)
                if ct is None:
                    ok = False
                    break
                params.append(f"{pname}: {ct}")
        if not ok:
            skipped += 1
            continue
        pdecl = ", ".join(params)
        decls.append(f"extern fn {fn}({pdecl}) -> {rclio}")
    return decls, skipped


def split_args(s: str) -> list[str]:
    out: list[str] = []
    cur: list[str] = []
    depth = 0
    for ch in s:
        if ch == "," and depth == 0:
            out.append("".join(cur).strip())
            cur = []
            continue
        if ch == "(":
            depth += 1
        elif ch == ")":
            depth -= 1
        cur.append(ch)
    if cur:
        out.append("".join(cur).strip())
    return [x for x in out if x]


def main() -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument("header")
    ap.add_argument("-o", "--out", default="include/raylib/raylib.clio")
    args = ap.parse_args()
    try:
        text = Path(args.header).read_text(encoding="utf-8", errors="replace")
    except OSError as e:
        print(e, file=sys.stderr)
        return 1

    structs = parse_structs(text)
    order = topo_structs(structs)
    struct_names = set(structs.keys())

    struct_blocks: list[str] = []
    for name in order:
        struct_blocks.append(f"pub struct {name} {{\n{structs[name]}\n}}")

    decls, skipped = parse_rlapi(text, struct_names)

    hdr = (
        f"' AUTO-GENERATED from raylib.h — run: python scripts/gen_raylib_clio.py <raylib.h> -o {args.out}\n"
        f"' {len(structs)} structs, {len(decls)} extern fn, {skipped} RLAPI lines skipped.\n"
        f"' # linkpath should point at the folder with raylib import library + raylib.h.\n"
        f'# link "raylib"\n'
        f'# linkpath "./raylib"\n\n'
    )
    body = "\n\n".join(struct_blocks)
    if struct_blocks:
        body += "\n\n"
    body += "\n\n".join(decls) + "\n"
    out = hdr + body

    out_path = Path(args.out)
    out_path.parent.mkdir(parents=True, exist_ok=True)
    out_path.write_text(out, encoding="utf-8", newline="\n")
    ex = Path("examples/libs/raylib/raylib.clio")
    if ex.resolve() != out_path.resolve():
        ex.parent.mkdir(parents=True, exist_ok=True)
        ex.write_text(out, encoding="utf-8", newline="\n")
        print(f"also wrote {ex}", file=sys.stderr)
    print(
        f"wrote {out_path}: structs={len(structs)} externs={len(decls)} skipped={skipped}",
        file=sys.stderr,
    )
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
