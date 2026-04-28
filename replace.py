import os
import re

def replace_arrow(exts):
    count = 0
    # regex matches a closing parenthesis, optional space, ->, optional space
    pattern = re.compile(r'\)\s*->\s*')
    
    for root, _, files in os.walk('.'):
        for file in files:
            if any(file.endswith(ext) for ext in exts):
                path = os.path.join(root, file)
                
                # skip this script itself
                if file == "replace.py" or file == "lex.go" or "next.go" in file or "token" in root:
                    continue
                    
                with open(path, 'r', encoding='utf-8') as f:
                    content = f.read()
                
                # Special replacement for markdown table row that got formatted:
                # e.g. `read_file("save.txt") -> str` in LANGUAGE.md
                # We can just run the regex
                new_content = pattern.sub(') ', content)
                
                if new_content != content:
                    with open(path, 'w', encoding='utf-8') as f:
                        f.write(new_content)
                    count += 1
                    print(f"Updated {path}")
    print(f"Total updated: {count} files.")

replace_arrow(['.kodae', '.go', '.md'])
