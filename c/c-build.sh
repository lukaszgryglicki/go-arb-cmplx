#!/bin/bash
gcc test_arb_str.c -o test_arb_str -I/usr/include/flint -L/usr/local/lib -lflint -lmpfr -lgmp && ./test_arb_str

