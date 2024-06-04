import sys
import os

inputFile = 'duplicates.txt'
if len(sys.argv) > 1:
    inputFile = sys.argv[1]

with open(inputFile, 'r') as f:

    for line in f:
        filepath = line.rstrip('\n')
        if os.path.isfile(filepath):
            print(f"Removing {filepath}\n")
            os.remove(filepath)
