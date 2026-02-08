#!/usr/bin/env python3
import re

with open('corfs/fs_test.go', 'r') as f:
    content = f.read()

# Pattern: testfs.File{ ... } that doesn't already have CloseFunc
# We need to add CloseFunc: func() error { return nil } before the closing }

def add_close_func(match):
    file_content = match.group(1)
    # Check if it already has CloseFunc
    if 'CloseFunc' in file_content:
        return match.group(0)  # Already has it, don't modify
    
    # Find the last closing brace and add CloseFunc before it
    # The file_content includes everything after "&testfs.File{" up to and including the closing "}"
    # We want to add CloseFunc before the final }
    
    # Find the last } that closes the struct
    lines = file_content.split('\n')
    result_lines = []
    
    for i, line in enumerate(lines):
        result_lines.append(line)
        # If this is the closing } line (just whitespace and })
        if i == len(lines) - 1 and re.match(r'^\s*}\s*$', line):
            # Insert CloseFunc before this line
            indent = re.match(r'^(\s*)}', line).group(1)
            result_lines.insert(-1, f'{indent}\tCloseFunc: func() error {{\n{indent}\t\treturn nil\n{indent}\t}},')
    
    return '&testfs.File{' + '\n'.join(result_lines)

# Match &testfs.File{ ... } including nested braces
# This is tricky with regex, so we'll use a simpler approach: find all &testfs.File{ and manually parse

lines = content.split('\n')
result = []
i = 0

while i < len(lines):
    line = lines[i]
    
    # Check if this line contains &testfs.File{
    if '&testfs.File{' in line:
        # Collect all lines until we find the matching }
        file_lines = [line]
        brace_count = line.count('{') - line.count('}')
        i += 1
        
        while i < len(lines) and brace_count > 0:
            file_lines.append(lines[i])
            brace_count += lines[i].count('{') - lines[i].count('}')
            i += 1
        
        # Check if CloseFunc exists in these lines
        file_content = '\n'.join(file_lines)
        if 'CloseFunc' not in file_content:
            # Add CloseFunc before the last }
            # Find the line with the final }
            for j in range(len(file_lines) - 1, -1, -1):
                if re.match(r'^\s*}\s*$', file_lines[j]):
                    indent = re.match(r'^(\s*)}', file_lines[j]).group(1)
                    # Insert CloseFunc before this line
                    file_lines.insert(j, f'{indent}\tCloseFunc: func() error {{')
                    file_lines.insert(j+1, f'{indent}\t\treturn nil')
                    file_lines.insert(j+2, f'{indent}\t}},')
                    break
        
        result.extend(file_lines)
    else:
        result.append(line)
        i += 1

with open('corfs/fs_test.go', 'w') as f:
    f.write('\n'.join(result))

print("Added CloseFunc to all testfs.File instances!")
