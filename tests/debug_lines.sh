#!/bin/bash

echo "=== Finding line numbers ==="
echo "Line with 'Successfully created a secret Gist':"
grep -n 'Successfully created a secret Gist' ../lib/gistlib.sh

echo ""
echo "Lines with 'gist_menu' in create_gist function:"
grep -n 'gist_menu' ../lib/gistlib.sh | head -10

echo ""
echo "Lines with 'return' after success message:"
awk '/Successfully created a secret Gist/{flag=1; line=NR} flag && /return/{print NR": "$0; if(NR-line < 30) exit}' ../lib/gistlib.sh
