const {
  Document, Packer, Paragraph, TextRun, Table, TableRow, TableCell,
  HeadingLevel, AlignmentType, BorderStyle, WidthType, ShadingType,
  LevelFormat, PageNumber, PageBreak, TabStopType, TabStopPosition,
  Header, Footer, ExternalHyperlink, PageNumberElement, NumberFormat
} = require('docx');
const fs = require('fs');
const path = require('path');

const BLUE     = "1a5fa8";
const DARKBLUE = "0d3d6b";
const LBLUE    = "dbeafe";
const GRAY     = "f3f4f6";
const DGRAY    = "374151";
const GREEN    = "d1fae5";
const DGREEN   = "065f46";
const RED      = "fee2e2";
const DRED     = "991b1b";
const AMBER    = "fef3c7";
const DAMBER   = "92400e";
const WHITE    = "ffffff";
const BORDER_C = "d1d5db";

const border  = { style: BorderStyle.SINGLE, size: 1, color: BORDER_C };
const borders = { top: border, bottom: border, left: border, right: border };
const noBorder = { style: BorderStyle.NONE, size: 0, color: "ffffff" };
const noBorders = { top: noBorder, bottom: noBorder, left: noBorder, right: noBorder };

function h1(text) {
  return new Paragraph({
    heading: HeadingLevel.HEADING_1,
    spacing: { before: 320, after: 160 },
    children: [new TextRun({ text, font: "Arial", size: 36, bold: true, color: DARKBLUE })]
  });
}

function h2(text) {
  return new Paragraph({
    heading: HeadingLevel.HEADING_2,
    spacing: { before: 280, after: 120 },
    children: [new TextRun({ text, font: "Arial", size: 28, bold: true, color: BLUE })]
  });
}

function h3(text) {
  return new Paragraph({
    heading: HeadingLevel.HEADING_3,
    spacing: { before: 200, after: 100 },
    children: [new TextRun({ text, font: "Arial", size: 24, bold: true, color: DGRAY })]
  });
}

function p(text, opts = {}) {
  return new Paragraph({
    spacing: { before: 80, after: 80 },
    children: [new TextRun({ text, font: "Arial", size: 22, color: DGRAY, ...opts })]
  });
}

function pRuns(runs) {
  return new Paragraph({
    spacing: { before: 80, after: 80 },
    children: runs.map(r => new TextRun({ font: "Arial", size: 22, color: DGRAY, ...r }))
  });
}

function code(text) {
  return new Paragraph({
    spacing: { before: 0, after: 0 },
    children: [new TextRun({ text, font: "Courier New", size: 20, color: "1e293b" })]
  });
}

function codeBlock(lines) {
  const cellBorder = { style: BorderStyle.SINGLE, size: 1, color: "cbd5e1" };
  const cellBorders = { top: cellBorder, bottom: cellBorder, left: cellBorder, right: cellBorder };
  return new Table({
    width: { size: 9360, type: WidthType.DXA },
    columnWidths: [9360],
    rows: [new TableRow({
      children: [new TableCell({
        borders: cellBorders,
        shading: { fill: "1e293b", type: ShadingType.CLEAR },
        margins: { top: 160, bottom: 160, left: 200, right: 200 },
        width: { size: 9360, type: WidthType.DXA },
        children: lines.map(l => new Paragraph({
          spacing: { before: 0, after: 0 },
          children: [new TextRun({ text: l, font: "Courier New", size: 19, color: "e2e8f0" })]
        }))
      })]
    })]
  });
}

function gap(sz = 160) {
  return new Paragraph({ spacing: { before: 0, after: sz }, children: [new TextRun("")] });
}

function bullet(text, indent = 0) {
  return new Paragraph({
    numbering: { reference: "bullets", level: indent },
    spacing: { before: 40, after: 40 },
    children: [new TextRun({ text, font: "Arial", size: 22, color: DGRAY })]
  });
}

function bulletRuns(runs, indent = 0) {
  return new Paragraph({
    numbering: { reference: "bullets", level: indent },
    spacing: { before: 40, after: 40 },
    children: runs.map(r => new TextRun({ font: "Arial", size: 22, color: DGRAY, ...r }))
  });
}

function twoCol(leftItems, rightItems, leftW = 4500, rightW = 4860) {
  return new Table({
    width: { size: 9360, type: WidthType.DXA },
    columnWidths: [leftW, rightW],
    rows: [new TableRow({
      children: [
        new TableCell({ borders: noBorders, width: { size: leftW, type: WidthType.DXA }, children: leftItems }),
        new TableCell({ borders: noBorders, width: { size: rightW, type: WidthType.DXA }, children: rightItems }),
      ]
    })]
  });
}

function infoBox(title, lines, color = LBLUE, textColor = DARKBLUE, borderColor = BLUE) {
  const brd = { style: BorderStyle.SINGLE, size: 2, color: borderColor };
  return new Table({
    width: { size: 9360, type: WidthType.DXA },
    columnWidths: [9360],
    rows: [new TableRow({
      children: [new TableCell({
        borders: { top: brd, bottom: brd, left: { style: BorderStyle.SINGLE, size: 12, color: borderColor }, right: brd },
        shading: { fill: color, type: ShadingType.CLEAR },
        margins: { top: 120, bottom: 120, left: 200, right: 200 },
        width: { size: 9360, type: WidthType.DXA },
        children: [
          title ? new Paragraph({ spacing: { before: 0, after: 60 }, children: [new TextRun({ text: title, font: "Arial", size: 22, bold: true, color: textColor })] }) : null,
          ...lines.map(l => new Paragraph({ spacing: { before: 0, after: 0 }, children: [new TextRun({ text: l, font: "Arial", size: 21, color: textColor })] }))
        ].filter(Boolean)
      })]
    })]
  });
}

function dataTable(headers, rows, colWidths) {
  const total = colWidths.reduce((a, b) => a + b, 0);
  const hdrBorder = { style: BorderStyle.SINGLE, size: 1, color: BORDER_C };
  const hdrBorders = { top: hdrBorder, bottom: hdrBorder, left: hdrBorder, right: hdrBorder };

  return new Table({
    width: { size: total, type: WidthType.DXA },
    columnWidths: colWidths,
    rows: [
      new TableRow({
        children: headers.map((h, i) => new TableCell({
          borders: hdrBorders,
          shading: { fill: BLUE, type: ShadingType.CLEAR },
          margins: { top: 80, bottom: 80, left: 120, right: 120 },
          width: { size: colWidths[i], type: WidthType.DXA },
          children: [new Paragraph({ children: [new TextRun({ text: h, font: "Arial", size: 20, bold: true, color: WHITE })] })]
        }))
      }),
      ...rows.map((row, ri) => new TableRow({
        children: row.map((cell, i) => new TableCell({
          borders: hdrBorders,
          shading: { fill: ri % 2 === 0 ? WHITE : GRAY, type: ShadingType.CLEAR },
          margins: { top: 60, bottom: 60, left: 120, right: 120 },
          width: { size: colWidths[i], type: WidthType.DXA },
          children: [new Paragraph({ children: [new TextRun({ text: cell, font: "Arial", size: 20, color: DGRAY })] })]
        }))
      }))
    ]
  });
}

function sectionDivider(label) {
  return new Paragraph({
    spacing: { before: 320, after: 160 },
    border: { bottom: { style: BorderStyle.SINGLE, size: 6, color: BLUE, space: 4 } },
    children: [new TextRun({ text: label, font: "Arial", size: 32, bold: true, color: DARKBLUE })]
  });
}

// ── BUILD DOCUMENT ────────────────────────────────────────────────────────────

const doc = new Document({
  numbering: {
    config: [
      { reference: "bullets", levels: [
        { level: 0, format: LevelFormat.BULLET, text: "•", alignment: AlignmentType.LEFT, style: { paragraph: { indent: { left: 720, hanging: 360 } } } },
        { level: 1, format: LevelFormat.BULLET, text: "◦", alignment: AlignmentType.LEFT, style: { paragraph: { indent: { left: 1080, hanging: 360 } } } },
      ]},
    ]
  },
  styles: {
    default: { document: { run: { font: "Arial", size: 22 } } },
    paragraphStyles: [
      { id: "Heading1", name: "Heading 1", basedOn: "Normal", next: "Normal", quickFormat: true,
        run: { size: 36, bold: true, font: "Arial", color: DARKBLUE },
        paragraph: { spacing: { before: 320, after: 160 }, outlineLevel: 0 } },
      { id: "Heading2", name: "Heading 2", basedOn: "Normal", next: "Normal", quickFormat: true,
        run: { size: 28, bold: true, font: "Arial", color: BLUE },
        paragraph: { spacing: { before: 280, after: 120 }, outlineLevel: 1 } },
      { id: "Heading3", name: "Heading 3", basedOn: "Normal", next: "Normal", quickFormat: true,
        run: { size: 24, bold: true, font: "Arial", color: DGRAY },
        paragraph: { spacing: { before: 200, after: 100 }, outlineLevel: 2 } },
    ]
  },
  sections: [{
    properties: {
      page: {
        size: { width: 12240, height: 15840 },
        margin: { top: 1440, right: 1440, bottom: 1440, left: 1440 }
      }
    },
    headers: {
      default: new Header({
        children: [new Paragraph({
          border: { bottom: { style: BorderStyle.SINGLE, size: 4, color: BLUE, space: 4 } },
          children: [
            new TextRun({ text: "Clio Language — Compiler Technical Specification", font: "Arial", size: 18, color: BLUE }),
            new TextRun({ text: "   |   v1.0   |   CONFIDENTIAL", font: "Arial", size: 18, color: "9ca3af" }),
          ]
        })]
      })
    },
    footers: {
      default: new Footer({
        children: [new Paragraph({
          border: { top: { style: BorderStyle.SINGLE, size: 4, color: BLUE, space: 4 } },
          tabStops: [{ type: TabStopType.RIGHT, position: 9360 }],
          children: [
            new TextRun({ text: "© Clio Language Project", font: "Arial", size: 18, color: "9ca3af" }),
            new TextRun({ text: "\tPage ", font: "Arial", size: 18, color: "9ca3af" }),
            new PageNumberElement(),
          ]
        })]
      })
    },
    children: [

      // ── COVER ──────────────────────────────────────────────────────────────
      new Paragraph({
        spacing: { before: 1200, after: 200 },
        children: [new TextRun({ text: "CLIO", font: "Arial", size: 96, bold: true, color: BLUE })]
      }),
      new Paragraph({
        spacing: { before: 0, after: 120 },
        children: [new TextRun({ text: "Compiler Technical Specification", font: "Arial", size: 40, color: DGRAY })]
      }),
      new Paragraph({
        spacing: { before: 0, after: 600 },
        children: [new TextRun({ text: "For the Go-based Clio → C Transpiler  |  v1.0", font: "Arial", size: 26, color: "9ca3af" })]
      }),
      infoBox("Language motto", [
        '"Write it like BASIC. Run it like C."',
        "",
        "Clio is a statically-typed systems language that transpiles to C.",
        "It gives programmers the readability of BASIC, the power of C,",
        "and the safety of modern languages — with zero runtime overhead.",
      ], LBLUE, DARKBLUE, BLUE),
      gap(400),

      // ── SECTION 1: OVERVIEW ────────────────────────────────────────────────
      sectionDivider("1.  Overview"),
      p("Clio is compiled by a Go program that reads Clio source files and outputs valid C99 code. That C code is then compiled by GCC or Clang to produce a native binary. This gives Clio C-equivalent speed, full access to every C library, and portability across Windows, macOS, and Linux with no extra work."),
      gap(80),

      h2("1.1  Compilation Pipeline"),
      codeBlock([
        "  Clio source (.clio)",
        "       |",
        "       v  [Go compiler]",
        "  Lexer  →  Parser  →  Type Checker  →  Code Generator",
        "       |",
        "       v",
        "  C source (.c)  +  clio_runtime.h",
        "       |",
        "       v  [GCC / Clang]",
        "  Native binary  (.exe / ELF / Mach-O)",
      ]),
      gap(100),

      h2("1.2  Design Goals"),
      bullet("Transpile to clean, readable C99 — not obfuscated output"),
      bullet("C-equivalent runtime performance — no GC pauses, no VM"),
      bullet("Every C library works out of the box via extern declarations"),
      bullet("Safe by default — bounds checked in debug, fast in release"),
      bullet("Simple enough to be understood by one programmer"),
      gap(160),

      // ── SECTION 2: LANGUAGE SYNTAX ─────────────────────────────────────────
      sectionDivider("2.  Language Syntax Reference"),

      h2("2.1  Comments"),
      p("Clio supports two comment styles:"),
      codeBlock([
        "  ' This is a single-line comment  (BASIC style — preferred)",
        "  -- This is also a single-line comment  (alternate)",
        "  /* This is a block comment */",
      ]),
      gap(100),

      h2("2.2  Variables"),
      codeBlock([
        "  let name = \"Alice\"         ' inferred str",
        "  let score: int = 0         ' explicit type",
        "  let ratio: float = 3.14",
        "  let alive: bool = true",
        "  const MAX_HEALTH = 100     ' constant — cannot be changed",
      ]),
      gap(100),

      h2("2.3  Types"),
      dataTable(
        ["Clio Type", "C Equivalent", "Notes"],
        [
          ["int",     "int64_t",   "Always 64-bit signed"],
          ["uint",    "uint64_t",  "Always 64-bit unsigned"],
          ["float",   "double",    "64-bit IEEE 754"],
          ["bool",    "bool",      "true / false"],
          ["str",     "clio_str",  "Real string — not char*"],
          ["byte",    "uint8_t",   "Raw 8-bit value"],
          ["ptr[T]",  "T*",        "Explicit pointer (C interop)"],
          ["[T]",     "struct",    "Dynamic growable array"],
          ["[T; N]",  "T[N]",      "Fixed-size array"],
          ["result[T]","struct",   "Value or error"],
          ["T?",      "struct",    "Optional — may be none"],
        ],
        [2000, 2200, 5160]
      ),
      gap(160),

      h2("2.4  Operators"),
      h3("Comparison"),
      codeBlock([
        "  ==   !=   <   >   <=   >=",
      ]),
      gap(80),
      h3("Logic  (BASIC-style keywords)"),
      codeBlock([
        "  AND    OR    NOT",
        "",
        "  if (alive AND score > 0) { ... }",
        "  if (dead OR quit) { ... }",
        "  if (NOT done) { ... }",
      ]),
      gap(80),
      h3("Math"),
      codeBlock([
        "  +   -   *   /   %",
      ]),
      gap(80),
      h3("Increment / Decrement"),
      codeBlock([
        "  x++    ' same as x = x + 1",
        "  x--    ' same as x = x - 1",
      ]),
      gap(80),
      h3("Compound Assignment"),
      codeBlock([
        "  x += 5    x -= 3    x *= 2    x /= 4    x %= 10",
      ]),
      gap(80),
      h3("Bitwise"),
      codeBlock([
        "  &   |   ^   ~   <<   >>",
        "",
        "  ' useful for color packing, flags, Raylib key masks",
        "  let color = (r << 16) | (g << 8) | b",
      ]),
      gap(160),

      h2("2.5  Strings"),
      p("Strings are a real type — not char*. They support concatenation, interpolation, and comparison directly."),
      codeBlock([
        "  let name = \"Alice\"",
        "  let bob  = \"Bob\"",
        "",
        "  ' Concatenation with +",
        "  print(\"Hello \" + name)",
        "",
        "  ' Kotlin-style interpolation — $ for variable, ${} for expression",
        "  print(\"Hello $name\")",
        "  print(\"Score: $score\")",
        "  print(\"Next level: ${score + 1}\")",
        "  print(\"Upper: ${name.upper()}\")",
        "",
        "  ' Comparison — just use ==",
        "  if (name == bob) { print(\"same!\") }",
        "  if (name == \"Alice\") { print(\"found!\") }",
      ]),
      gap(100),

      h2("2.6  Control Flow"),
      codeBlock([
        "  ' If / else if / else",
        "  if (score > 100) {",
        "      print(\"winner\")",
        "  } else if (score > 50) {",
        "      print(\"close\")",
        "  } else {",
        "      print(\"keep trying\")",
        "  }",
        "",
        "  ' While loop",
        "  while (health > 0) {",
        "      health -= damage",
        "  }",
        "",
        "  ' For range  (0 up to but not including 10)",
        "  for (i in 0..10) {",
        "      print(\"$i\")",
        "  }",
        "",
        "  ' For collection",
        "  for (item in inventory) {",
        "      print(\"$item\")",
        "  }",
        "",
        "  ' Infinite loop with break",
        "  loop {",
        "      if (done) { break }",
        "  }",
      ]),
      gap(100),

      h2("2.7  Functions"),
      codeBlock([
        "  fn add(a: int, b: int) -> int {",
        "      return a + b",
        "  }",
        "",
        "  ' Multiple return values",
        "  fn divmod(a: int, b: int) -> (int, int) {",
        "      return a / b, a % b",
        "  }",
        "",
        "  ' String return with interpolation",
        "  fn greet(name: str) -> str {",
        "      return \"Hello, $name!\"",
        "  }",
      ]),
      gap(100),

      h2("2.8  Structs and Methods"),
      codeBlock([
        "  struct Player {",
        "      name:   str",
        "      health: int",
        "      x:      float",
        "      y:      float",
        "  }",
        "",
        "  ' Method — first param is always self",
        "  fn Player.hurt(self, amount: int) {",
        "      self.health -= amount",
        "      if (self.health <= 0) {",
        "          print(\"$self.name has died!\")",
        "      }",
        "  }",
        "",
        "  fn Player.is_alive(self) -> bool {",
        "      return self.health > 0",
        "  }",
        "",
        "  ' Create and use",
        "  let p = Player { name: \"Hero\", health: 100, x: 0.0, y: 0.0 }",
        "  p.hurt(25)",
        "  if (p.is_alive()) { print(\"Still fighting!\") }",
      ]),
      gap(100),

      h2("2.9  Enums and Match"),
      codeBlock([
        "  enum State { Menu, Playing, Paused, GameOver }",
        "",
        "  let state: State = State.Playing",
        "",
        "  match (state) {",
        "      State.Menu     => { show_menu() }",
        "      State.Playing  => { update_game() }",
        "      State.Paused   => { draw_pause_screen() }",
        "      State.GameOver => { print(\"Game Over!\") }",
        "  }",
      ]),
      gap(100),

      h2("2.10  Error Handling"),
      codeBlock([
        "  ' Function that can fail returns result[T]",
        "  fn read_file(path: str) -> result[str] { ... }",
        "",
        "  ' Handle the error with catch",
        "  let data = read_file(\"save.dat\") catch (err) {",
        "      print(\"Could not load: $err\")",
        "      return",
        "  }",
        "",
        "  ' Propagate error up to caller with ?",
        "  let data = read_file(\"save.dat\")?",
      ]),
      gap(100),

      h2("2.11  Arrays"),
      codeBlock([
        "  ' Fixed array",
        "  let nums: [int; 5] = [1, 2, 3, 4, 5]",
        "",
        "  ' Dynamic array",
        "  let items: [str] = []",
        "  items.push(\"sword\")",
        "  items.push(\"shield\")",
        "  print(\"Count: ${items.len}\")",
        "  print(\"First: ${items[0]}\")",
        "  items.pop()",
        "  items.sort()",
        "  items.clear()",
      ]),
      gap(100),

      h2("2.12  Built-in Functions"),
      dataTable(
        ["Function", "Description", "Example"],
        [
          ["print(v)",          "Print any value + newline",      "print(\"Hi $name\")"],
          ["input(prompt)",     "Read a line from the user",      "let s = input(\"Name: \")"],
          ["input_int(prompt)", "Read an integer from the user",  "let n = input_int(\"Age: \")"],
          ["random(lo, hi)",    "Random int in range",            "let n = random(1, 6)"],
          ["random_float()",    "Random float 0.0..1.0",          "let f = random_float()"],
          ["int(x)",            "Convert to int",                 "int(3.7)  →  3"],
          ["float(x)",          "Convert to float",               "float(5)  →  5.0"],
          ["str(x)",            "Convert anything to str",        "str(42)   →  \"42\""],
          ["abs(x)",            "Absolute value",                 "abs(-5)   →  5"],
          ["min(a,b)",          "Smaller of two values",          "min(3, 7) →  3"],
          ["max(a,b)",          "Larger of two values",           "max(3, 7) →  7"],
          ["clamp(x,lo,hi)",    "Clamp x between lo and hi",     "clamp(hp, 0, 100)"],
          ["sqrt(x)",           "Square root",                    "sqrt(9.0) →  3.0"],
          ["pow(x, y)",         "x to the power y",              "pow(2,10) →  1024"],
          ["floor(x)",          "Round down",                     "floor(3.9)→  3.0"],
          ["ceil(x)",           "Round up",                       "ceil(3.1) →  4.0"],
          ["round(x)",          "Round to nearest",               "round(3.5)→  4.0"],
          ["sin(x)",            "Sine (radians)",                 "sin(0.0)  →  0.0"],
          ["cos(x)",            "Cosine (radians)",               "cos(0.0)  →  1.0"],
          ["time()",            "Seconds since program start",    "let t = time()"],
          ["sleep(secs)",       "Pause execution",                "sleep(1.0)"],
          ["exit(code)",        "Quit the program",               "exit(0)"],
          ["assert(cond, msg)", "Crash with message if false",    "assert(hp>=0, \"bad hp\")"],
          ["debug(val)",        "Print any value for debugging",  "debug(player)"],
          ["read_file(path)",   "Read file → result[str]",        "read_file(\"data.txt\")"],
          ["write_file(p, s)",  "Write string to file",           "write_file(\"out.txt\", s)"],
          ["file_exists(path)", "Check if file exists",           "file_exists(\"save.dat\")"],
          ["clear_screen()",    "Clear terminal output",          "clear_screen()"],
          ["args()",            "Command-line args as [str]",     "let a = args()"],
        ],
        [2200, 3000, 4160]
      ),
      gap(160),

      h2("2.13  String Methods"),
      codeBlock([
        "  name.len              ' length as int",
        "  name.upper()          ' \"ALICE\"",
        "  name.lower()          ' \"alice\"",
        "  name.trim()           ' remove whitespace from ends",
        "  name.contains(\"li\")   ' true / false",
        "  name.starts(\"Al\")     ' true / false",
        "  name.ends(\"ce\")       ' true / false",
        "  name.replace(\"l\",\"r\") ' returns new string",
        "  name.split(\",\")       ' returns [str]",
        "  name.slice(0, 3)      ' substring from index 0 to 3",
      ]),
      gap(100),

      h2("2.14  Modules"),
      codeBlock([
        "  ' math.clio",
        "  module math",
        "",
        "  pub fn square(x: int) -> int {",
        "      return x * x",
        "  }",
        "",
        "  ' main.clio",
        "  use math",
        "",
        "  fn main() {",
        "      let n = math.square(5)",
        "      print(\"$n\")",
        "  }",
      ]),
      gap(100),

      h2("2.15  C Interop"),
      codeBlock([
        "  ' Declare an external C function",
        "  extern fn printf(fmt: ptr[byte], ...) -> int",
        "",
        "  ' Link a C library",
        "  #link \"raylib\"",
        "  extern fn InitWindow(w: int, h: int, title: ptr[byte])",
        "  extern fn CloseWindow()",
        "  extern fn WindowShouldClose() -> bool",
        "",
        "  ' Defer runs cleanup at end of scope (like Go)",
        "  fn main() {",
        "      InitWindow(800, 600, \"My Game\")",
        "      defer CloseWindow()",
        "      loop {",
        "          if (WindowShouldClose()) { break }",
        "      }",
        "  }",
      ]),
      gap(160),

      // ── SECTION 3: GO COMPILER ARCHITECTURE ───────────────────────────────
      sectionDivider("3.  Go Compiler Architecture"),
      p("The Clio compiler is a single Go program split into clean packages. Each package handles exactly one stage of compilation."),
      gap(80),

      h2("3.1  Package Structure"),
      codeBlock([
        "  clio/",
        "  ├── main.go             CLI entry point: run, build, check, bind",
        "  ├── lexer/",
        "  │   └── lexer.go       ' text → []Token",
        "  ├── parser/",
        "  │   └── parser.go      ' []Token → AST",
        "  ├── ast/",
        "  │   └── nodes.go       ' AST node type definitions",
        "  ├── resolver/",
        "  │   └── resolver.go    ' name resolution, scope checking",
        "  ├── typechecker/",
        "  │   └── checker.go     ' type inference and verification",
        "  ├── codegen/",
        "  │   └── codegen.go     ' AST → C source string",
        "  ├── runtime/",
        "  │   └── clio_runtime.h ' C header bundled into every build",
        "  └── binder/",
        "      └── binder.go      ' parse .h files → .clio extern declarations",
      ]),
      gap(160),

      h2("3.2  Stage 1 — Lexer"),
      p("Input: raw Clio source string. Output: slice of Token structs."),
      gap(60),
      infoBox("Key responsibilities", [
        "• Recognize all keywords: fn let const if else while for in loop break continue struct enum match use module pub extern defer",
        "• Handle both comment styles:  '  and  --",
        "• Lex string interpolation: split \"Hello $name!\" into [STR_START, IDENT, STR_END]",
        "• Lex $var (bare) and ${expr} (expression) interpolation forms",
        "• Track line and column numbers for all tokens (needed for error messages)",
        "• Produce clean error: unexpected character '§' at line 4, col 12",
      ], GRAY, DGRAY, BORDER_C),
      gap(100),
      h3("Token kinds to implement"),
      dataTable(
        ["Category", "Tokens"],
        [
          ["Literals",     "INT  FLOAT  STR_PLAIN  STR_START  STR_MID  STR_END  BOOL"],
          ["Keywords",     "FN  LET  CONST  RETURN  IF  ELSE  WHILE  FOR  IN  LOOP  BREAK  CONTINUE"],
          ["Keywords",     "STRUCT  ENUM  MATCH  USE  MODULE  PUB  EXTERN  DEFER  AND  OR  NOT"],
          ["Operators",    "PLUS  MINUS  STAR  SLASH  PERCENT  EQ  EQEQ  NEQ  LT  GT  LTE  GTE"],
          ["Operators",    "PLUSEQ  MINUSEQ  STAREQ  SLASHEQ  PERCENTEQ  PLUSPLUS  MINUSMINUS"],
          ["Operators",    "AMP  PIPE  CARET  TILDE  LSHIFT  RSHIFT  ARROW  DOTDOT  QUEST"],
          ["Delimiters",   "LPAREN  RPAREN  LBRACE  RBRACE  LBRACK  RBRACK  COMMA  COLON  DOT  SEMI  HASH"],
          ["Special",      "IDENT  EOF"],
        ],
        [2000, 7360]
      ),
      gap(160),

      h2("3.3  Stage 2 — Parser"),
      p("Input: []Token. Output: Program AST. Uses recursive descent — one function per grammar rule."),
      gap(60),
      infoBox("Key responsibilities", [
        "• Parse all top-level declarations: fn, struct, enum, use, module, extern, const, #link",
        "• Parse all statements: let, const, if, while, for, loop, match, return, break, continue, defer",
        "• Parse all expressions with correct precedence (Pratt parsing recommended)",
        "• Handle string interpolation — rebuild StrInterp AST node from lexer tokens",
        "• Produce helpful errors: expected '{' after if condition, got 'for' at line 7",
        "• Semicolons are optional — newlines end statements like in Go",
      ], GRAY, DGRAY, BORDER_C),
      gap(100),
      h3("Operator precedence table (lowest to highest)"),
      dataTable(
        ["Level", "Operators", "Associativity"],
        [
          ["1 (lowest)", "OR",                   "left"],
          ["2",          "AND",                  "left"],
          ["3",          "NOT",                  "right (unary)"],
          ["4",          "==  !=  <  >  <=  >=", "left"],
          ["5",          "+  -",                 "left"],
          ["6",          "*  /  %",              "left"],
          ["7",          "unary -  ~",           "right"],
          ["8 (highest)","++  --  . [] ()",      "left (postfix)"],
        ],
        [1400, 5160, 2800]
      ),
      gap(160),

      h2("3.4  Stage 3 — Resolver"),
      p("Input: Program AST. Output: same AST, with scopes attached. Catches name errors before type checking."),
      gap(60),
      infoBox("Key responsibilities", [
        "• Build a scope tree — each { } block gets its own scope",
        "• Verify every name is declared before use",
        "• Resolve method calls: Player.hurt → fn__Player__hurt in the symbol table",
        "• Verify break and continue are inside a loop",
        "• Verify return is inside a function",
        "• Check for duplicate variable names in the same scope",
        "• Track which module each name comes from",
      ], GRAY, DGRAY, BORDER_C),
      gap(160),

      h2("3.5  Stage 4 — Type Checker"),
      p("Input: resolved AST. Output: same AST with types annotated on every expression node."),
      gap(60),
      infoBox("Key responsibilities", [
        "• Infer types for let declarations without explicit annotation",
        "• Verify both sides of binary operators match (or can be coerced)",
        "• Verify function call argument types match parameter types",
        "• Verify return type matches function signature",
        "• Verify struct field types on struct literals",
        "• Check match statements are exhaustive for enums",
        "• Handle str + str → str (concatenation), int + float → float (coercion)",
        "• Verify result[T] values are handled before use (catch or ?)",
      ], GRAY, DGRAY, BORDER_C),
      gap(160),

      h2("3.6  Stage 5 — Code Generator"),
      p("Input: typed AST. Output: a single C99 source file as a string."),
      gap(60),
      infoBox("Key responsibilities", [
        "• Emit #include \"clio_runtime.h\" at the top of every file",
        "• Emit -link flags as comments for the driver to read",
        "• Translate every Clio type to its C equivalent (see table in section 2.3)",
        "• Translate string interpolation to a series of clio_str_concat() calls",
        "• Translate methods: Player.hurt → void clio_Player__hurt(Player* self, int64_t amount)",
        "• Translate match/enum to C switch statements",
        "• Translate defer using a cleanup label + goto at end of function",
        "• Translate result[T] to a C struct with .value, .err, .ok fields",
        "• Emit bounds-check calls in debug mode, remove them in release mode",
      ], GRAY, DGRAY, BORDER_C),
      gap(100),
      h3("Example: Clio → C translation"),
      twoCol(
        [
          p("Clio input:", { bold: true }),
          codeBlock([
            "let name = \"Alice\"",
            "let score = 42",
            "print(\"Hi $name: ${score+1}\")",
          ]),
        ],
        [
          p("C output:", { bold: true }),
          codeBlock([
            "clio_str name =",
            "  clio_str_from(\"Alice\");",
            "int64_t score = 42;",
            "clio_print(clio_str_concat(",
            "  clio_str_concat(",
            "    clio_str_from(\"Hi \"),",
            "    name),",
            "  clio_str_concat(",
            "    clio_str_from(\": \"),",
            "    clio_int_to_str(score+1))));",
          ]),
        ]
      ),
      gap(160),

      // ── SECTION 4: RUNTIME ─────────────────────────────────────────────────
      sectionDivider("4.  The Clio Runtime (clio_runtime.h)"),
      p("The entire Clio runtime is a single C header file bundled with every build. It has no external dependencies beyond the C standard library. Keep it small — the goal is under 500 lines."),
      gap(80),

      h2("4.1  String Type"),
      codeBlock([
        "  typedef struct {",
        "      const char* data;   /* pointer to string bytes */",
        "      int64_t     len;    /* length (not including \\0) */",
        "  } clio_str;",
        "",
        "  clio_str clio_str_from(const char* s);",
        "  clio_str clio_str_concat(clio_str a, clio_str b);",
        "  bool     clio_str_eq(clio_str a, clio_str b);",
        "  clio_str clio_str_upper(clio_str s);",
        "  clio_str clio_str_lower(clio_str s);",
        "  clio_str clio_str_trim(clio_str s);",
        "  bool     clio_str_contains(clio_str s, clio_str sub);",
        "  bool     clio_str_starts(clio_str s, clio_str prefix);",
        "  bool     clio_str_ends(clio_str s, clio_str suffix);",
        "  clio_str clio_str_replace(clio_str s, clio_str from, clio_str to);",
        "  clio_str clio_str_slice(clio_str s, int64_t start, int64_t end);",
        "  clio_str clio_int_to_str(int64_t n);",
        "  clio_str clio_float_to_str(double n);",
        "  clio_str clio_bool_to_str(bool b);",
      ]),
      gap(100),

      h2("4.2  Dynamic Array (generic macro)"),
      codeBlock([
        "  /* Declare a typed array: */",
        "  #define CLIO_ARR(T)  struct { T* data; int64_t len; int64_t cap; }",
        "",
        "  /* Push an element (doubles capacity when full): */",
        "  #define CLIO_ARR_PUSH(arr, val)  \\",
        "      if ((arr).len >= (arr).cap) { \\",
        "          (arr).cap = (arr).cap ? (arr).cap * 2 : 8; \\",
        "          (arr).data = realloc((arr).data, sizeof(*(arr).data) * (arr).cap); \\",
        "      } \\",
        "      (arr).data[(arr).len++] = (val);",
        "",
        "  /* Pop last element: */",
        "  #define CLIO_ARR_POP(arr)  ((arr).data[--(arr).len])",
      ]),
      gap(100),

      h2("4.3  Result Type (error handling)"),
      codeBlock([
        "  /* result[str] becomes: */",
        "  typedef struct {",
        "      clio_str value;",
        "      clio_str err;",
        "      bool     ok;",
        "  } clio_result_clio_str;",
        "",
        "  /* ok() and err() constructors: */",
        "  #define CLIO_OK(T, val)   ((T){ .value=(val), .ok=true })",
        "  #define CLIO_ERR(T, msg)  ((T){ .err=clio_str_from(msg), .ok=false })",
      ]),
      gap(100),

      h2("4.4  Memory Model"),
      infoBox("Rules for the code generator to follow", [
        "• All local variables → stack allocated (C locals)",
        "• Dynamic arrays [T] → heap via malloc/realloc, freed when scope exits (emit free() at scope end)",
        "• Strings are immutable — safe to share the pointer without copying",
        "• String concatenation always allocates a new string on the heap",
        "• ptr[T] → raw C pointer — the programmer manages it manually",
        "• Structs passed by value under 64 bytes, by pointer above (for performance)",
        "• In debug mode: emit bounds checks on all array accesses",
        "• In release mode: omit bounds checks for maximum speed",
      ], AMBER, DAMBER, "d97706"),
      gap(160),

      // ── SECTION 5: BUILD SYSTEM ────────────────────────────────────────────
      sectionDivider("5.  Build System and CLI"),

      h2("5.1  CLI Commands"),
      dataTable(
        ["Command", "Description"],
        [
          ["clio run game.clio",          "Compile and immediately run the program"],
          ["clio build game.clio",        "Compile to a binary (debug mode)"],
          ["clio build --release game.clio", "Compile with full optimizations"],
          ["clio check game.clio",        "Type-check only — no output produced"],
          ["clio bind raylib.h",          "Generate .clio extern declarations from a C header"],
          ["clio version",                "Print compiler version"],
        ],
        [4000, 5360]
      ),
      gap(100),

      h2("5.2  GCC / Clang Flags"),
      codeBlock([
        "  /* Debug build: */",
        "  gcc -O0 -g -fsanitize=address,undefined -o output output.c -lraylib",
        "",
        "  /* Release build: */",
        "  gcc -O3 -march=native -flto -o output output.c -lraylib",
        "",
        "  /* The compiler reads #link directives from the generated C and",
        "     appends -l<name> flags automatically. */",
      ]),
      gap(160),

      // ── SECTION 6: C LIBRARY INTEROP ──────────────────────────────────────
      sectionDivider("6.  C Library Interop"),

      h2("6.1  Using any C library from Clio"),
      p("Any C library works. The programmer declares the functions with extern and links the library with #link:"),
      codeBlock([
        "  #link \"raylib\"",
        "",
        "  extern fn InitWindow(w: int, h: int, title: ptr[byte])",
        "  extern fn BeginDrawing()",
        "  extern fn EndDrawing()",
        "  extern fn ClearBackground(color: uint)",
        "  extern fn DrawText(text: ptr[byte], x: int, y: int, size: int, color: uint)",
        "  extern fn WindowShouldClose() -> bool",
        "  extern fn CloseWindow()",
        "",
        "  fn main() {",
        "      InitWindow(800, 600, \"Hello Raylib\")",
        "      defer CloseWindow()",
        "      loop {",
        "          if (WindowShouldClose()) { break }",
        "          BeginDrawing()",
        "          ClearBackground(0x181818FF)",
        "          DrawText(\"Hello from Clio!\", 300, 280, 24, 0xFFFFFFFF)",
        "          EndDrawing()",
        "      }",
        "  }",
      ]),
      gap(100),

      h2("6.2  The clio bind command"),
      p("The binder reads a C header file and auto-generates extern declarations so the programmer doesn't have to write them by hand:"),
      codeBlock([
        "  $ clio bind raylib.h > raylib.clio",
        "",
        "  ' Generated output (excerpt):",
        "  #link \"raylib\"",
        "  extern fn InitWindow(width: int, height: int, title: ptr[byte])",
        "  extern fn DrawRectangle(x: int, y: int, w: int, h: int, color: uint)",
        "  extern fn GetFrameTime() -> float",
        "  ' ... (all functions from raylib.h)",
      ]),
      gap(60),
      infoBox("Implementation note", [
        "The binder is a simple C header parser in Go. It does NOT need to be a full C parser.",
        "It only needs to handle function declarations (return type, name, parameters).",
        "typedef, #define constants, and struct declarations are out of scope for v1.",
        "Use Go's regexp or a simple line-by-line parser — this does not need to be perfect.",
      ], GREEN, DGREEN, "059669"),
      gap(160),

      // ── SECTION 7: PERFORMANCE GUIDE ──────────────────────────────────────
      sectionDivider("7.  Performance Guide"),
      p("Clio compiles to C, so it inherits C performance directly. Here are the specific rules to follow to ensure the output is as fast as hand-written C++."),
      gap(80),

      bullet("Pass structs ≤ 64 bytes by value, larger structs by pointer — avoids unnecessary copying"),
      bullet("Emit static inline for small functions (len < 10 lines) — the C compiler will inline them"),
      bullet("Dynamic arrays use capacity doubling — never reallocate one element at a time"),
      bullet("String concatenation allocates once per concat — the code generator should not chain mallocs"),
      bullet("In release mode, remove ALL bounds checks — they are only for debug safety"),
      bullet("Use -O3 -march=native -flto for release — this alone matches C++ performance"),
      bullet("Enums compile to C enum (int) — match compiles to switch — both are zero-cost"),
      bullet("Methods compile to plain C functions with a self pointer — no vtable, no overhead"),
      gap(160),

      // ── SECTION 8: ERROR MESSAGES ─────────────────────────────────────────
      sectionDivider("8.  Error Message Guidelines"),
      p("Good error messages are a core feature of Clio. Every error must tell the programmer exactly what went wrong and where."),
      gap(80),

      h2("8.1  Format"),
      codeBlock([
        "  error: <what went wrong>",
        "    --> filename.clio:line:col",
        "     |",
        "  42 |  if (name = \"Alice\") {",
        "     |         ^ did you mean == instead of = ?",
      ]),
      gap(80),

      h2("8.2  Required error messages"),
      dataTable(
        ["Situation", "Message to show"],
        [
          ["Unexpected character",       "unexpected character '§' — is this a typo?"],
          ["Unterminated string",        "string opened on line 4 was never closed with \""],
          ["Assignment in condition",    "did you mean == instead of = here?"],
          ["Unknown variable",           "name 'scroe' not found — did you mean 'score'?"],
          ["Type mismatch",              "expected int but got str — use int(x) to convert"],
          ["Missing return",             "function 'add' must return int but has no return statement"],
          ["Break outside loop",         "break can only be used inside a loop"],
          ["Non-exhaustive match",       "match on State is missing: State.Paused, State.GameOver"],
          ["Wrong arg count",            "fn greet() takes 1 argument but 2 were given"],
          ["Result not handled",         "read_file() returns result[str] — use catch or ? to handle it"],
        ],
        [3200, 6160]
      ),
      gap(160),

      // ── SECTION 9: BUILD ORDER ─────────────────────────────────────────────
      sectionDivider("9.  Recommended Build Order"),
      p("Build the compiler in this exact order. Do not skip ahead — each stage depends on the previous one working correctly."),
      gap(80),

      infoBox("Phase 1 — Get any output at all", [
        "1. Write the Lexer. Test it on 'let x = 42' and print the tokens.",
        "2. Write the Parser. Test it produces an AST for a hello world program.",
        "3. Write a minimal Codegen that emits C for: let, print, if, while, fn.",
        "4. Write the Driver that calls GCC on the output.",
        "5. Milestone: 'clio run hello.clio' prints Hello, World!",
      ], GREEN, DGREEN, "059669"),
      gap(80),
      infoBox("Phase 2 — Make it useful", [
        "6. Add structs, methods, enums, match.",
        "7. Add string interpolation ($var and ${expr}).",
        "8. Add dynamic arrays with push/pop/len.",
        "9. Add result[T] error handling (catch and ?).",
        "10. Milestone: can write a simple text adventure game.",
      ], LBLUE, DARKBLUE, BLUE),
      gap(80),
      infoBox("Phase 3 — Make it production-ready", [
        "11. Write the Resolver (name checking).",
        "12. Write the Type Checker.",
        "13. Write the Binder (clio bind raylib.h).",
        "14. Add defer, ++/--, +=/-=/*=, compound operators.",
        "15. Add the module system (use / module / pub).",
        "16. Add helpful error messages with line/col pointers.",
        "17. Milestone: can build a full Raylib game in Clio.",
      ], AMBER, DAMBER, "d97706"),
      gap(160),

      // ── SECTION 10: QUICK REFERENCE ───────────────────────────────────────
      sectionDivider("10.  Quick Reference Card"),
      p("Cut this out and keep it next to your keyboard."),
      gap(80),
      codeBlock([
        "  ' ── VARIABLES ─────────────────────────────────────────────────",
        "  let x = 10          let name: str = \"Alice\"       const MAX = 100",
        "",
        "  ' ── STRINGS ──────────────────────────────────────────────────",
        "  print(\"Hi $name\")           print(\"Val: ${x + 1}\")",
        "  print(\"Hello \" + name)      if (name == bob) { ... }",
        "",
        "  ' ── OPERATORS ────────────────────────────────────────────────",
        "  x++  x--  x += 5  x -= 3  x *= 2  x /= 4  x %= 10",
        "  AND  OR  NOT    ==  !=  <  >  <=  >=    &  |  ^  ~  <<  >>",
        "",
        "  ' ── CONTROL FLOW ─────────────────────────────────────────────",
        "  if (x > 0) { } else if (x < 0) { } else { }",
        "  while (alive AND score < MAX) { }",
        "  for (i in 0..10) { }    for (item in list) { }",
        "  loop { if (done) { break } }",
        "",
        "  ' ── FUNCTIONS ────────────────────────────────────────────────",
        "  fn add(a: int, b: int) -> int { return a + b }",
        "",
        "  ' ── STRUCTS / METHODS ────────────────────────────────────────",
        "  struct Point { x: float  y: float }",
        "  fn Point.move(self, dx: float, dy: float) { self.x += dx }",
        "",
        "  ' ── ENUMS / MATCH ────────────────────────────────────────────",
        "  enum State { Menu, Playing, GameOver }",
        "  match (state) { State.Menu => { } State.Playing => { } }",
        "",
        "  ' ── ERROR HANDLING ───────────────────────────────────────────",
        "  let data = read_file(\"x.txt\") catch (err) { print(\"$err\") }",
        "",
        "  ' ── C INTEROP ────────────────────────────────────────────────",
        "  #link \"raylib\"",
        "  extern fn InitWindow(w: int, h: int, title: ptr[byte])",
        "  defer CloseWindow()",
      ]),
      gap(200),
      new Paragraph({
        alignment: AlignmentType.CENTER,
        children: [new TextRun({ text: "End of Specification — Clio v1.0", font: "Arial", size: 20, color: "9ca3af" })]
      }),
    ]
  }]
});

Packer.toBuffer(doc).then(buf => {
  fs.writeFileSync(path.join(__dirname, 'Clio_Compiler_Spec.docx'), buf);
  console.log('done');
});