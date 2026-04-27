/**
 * One-time helper: read this chat's user message (docx build script) from
 * agent transcript JSONL and write build-spec.js with a proper output path.
 * Usage: node extract-and-write-build-spec.cjs <path-to-agent.jsonl>
 */
/* eslint-disable no-console */
const fs = require('fs');
const path = require('path');

const transcriptPath = process.argv[2]
  || path.join(
    'C:/Users/djcra/.cursor/projects/c-Users-djcra-my-shit-Desktop-Nova/agent-transcripts/4cb0a405-e1ce-4866-ad01-398b6f54fb93/4cb0a405-e1ce-4866-ad01-398b6f54fb93.jsonl',
  );

const line1 = fs.readFileSync(transcriptPath, 'utf8').split(/\r?\n/)[0];
const j = JSON.parse(line1);
const t = j.message?.content?.[0]?.text;
if (typeof t !== 'string' || t.length < 200) {
  throw new Error('Transcript first line: expected user message with docx script');
}

const m = t.match(/<user_query>\s*([\s\S]*?)\s*<\/user_query>\s*$/i)
  || t.match(/<user_query>\s*([\s\S]*?)\s*<\/user_query>/);
if (!m) {
  throw new Error('No <user_query> block in transcript text');
}
let code = m[1].trim();

if (!code.startsWith('const {')) {
  throw new Error('Extracted block does not start with const {');
}

if (!code.includes("require('fs')") && !code.includes('require("fs")')) {
  throw new Error('Unexpected: fs require missing');
}

if (!code.includes("fs.writeFileSync('/mnt/user-data/outputs/Clio_Compiler_Spec.docx'")) {
  code = code.replace(
    /fs\.writeFileSync\([^\n]+\)/,
    "fs.writeFileSync(path.join(__dirname, 'Clio_Compiler_Spec.docx')",
  );
} else {
  code = code.replace(
    "fs.writeFileSync('/mnt/user-data/outputs/Clio_Compiler_Spec.docx', buf);",
    "fs.writeFileSync(path.join(__dirname, 'Clio_Compiler_Spec.docx'), buf);",
  );
}

if (code.indexOf("const path = require('path');") < 0 && !code.includes('const path = require("path")')) {
  code = code.replace("const fs = require('fs');", "const fs = require('fs');\nconst path = require('path');");
}

const out = path.join(__dirname, 'build-spec.js');
fs.writeFileSync(out, code, 'utf8');
console.log('Wrote', out, '(', code.length, 'bytes )');
