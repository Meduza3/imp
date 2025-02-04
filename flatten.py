#!/usr/bin/env python3
import sys

def flatten_text(text):
    # Split on any whitespace and rejoin with a single space
    tokens = text.split()
    return " ".join(tokens)

if __name__ == "__main__":
    text = sys.stdin.read()
    print(flatten_text(text))
