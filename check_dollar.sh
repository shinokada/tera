#!/usr/bin/env bash

# Extract line 96 and show its hex dump
sed -n '96p' /Users/shinichiokada/Bash/tera/lib/gistlib.sh | od -A n -t x1
