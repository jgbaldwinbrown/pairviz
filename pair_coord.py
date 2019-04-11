#!/usr/bin/env python3

import sys
import argparse

class Entry(object):
    def __init__:
        self.subject_chrom = ""
        self.query_chrom = ""
        self.subject_start = None
        self.query_start = None
        self.current_subject = None
        self.current_query = None
        self.conv = {"query": [], "subject": []}

def add_to_entry(entry):
    entry.conv["query"] = entry.current_query
    entry.conv["subject"] = entry.current_subject

def start_entry(alignment_line, chrominfo):
    sl = alignment_line.split()
    entry = Entry()
    entry.subject_chrom = chrominfo[0]
    entry.query_chrom = chrominfo[1]
    entry.subject_start = int(sl[0])
    entry.subject_end = int(sl[1])
    entry.query_start = int(sl[2])
    entry.query_end = int(sl[3])
    entry.current_subject = entry.subject_start
    entry.current_query = entry.query_start
    add_to_entry(entry)
    return(entry)

def increment_subject(entry):
    if entry.subject_start < entry.subject_end:
        entry.current_subject += 1
    else:
        entry.current_subject -= 1

def increment_query(entry):
    if entry.query_start < entry.query_end:
        entry.current_query += 1
    else:
        entry.current_query -= 1

def update_entry(entry, line):
    change = int(line)
    for i in range(change-1):
        increment_subject(entry)
        increment_query(entry)
        add_to_entry(entry)
    elif change > 0:
        increment_subject(entry)
    elif change < 0:
        increment_query(entry)
    else:
        exit("This should be impossible!")

def add_entry(entry, coord_conv)
    for querypos, subjectpos in zip(entry.conv["query"], entry.conv["subject"]):
        coord_conv[(entry.query_chrom, querypos)] = (entry.subject_chrom, subjectpos)

def get_coord_conv(mummer_delta):
    coord_conv = {}
    entry = None
    for i,l in enumerate(open(mummer_delta, "r")):
        l=l.rstrip('\n')
        sl = l.split('\t')
        if i==0:
            files = l
        elif i==1:
            program = l
        elif l[0] == ">":
            chrominfo = l.lstrip(">").split()[:2]
        elif len(sl) > 1:
            entry = start_entry(l, chrominfo)
        elif int(l) == 0:
            add_entry(entry, coord_conv)
        else:
            update_entry(entry, l)
    return(coord_conv)

def conv_file(inconn, coord_conv, mummer_query):
    #for inconn in inconns:
    #    for l in inconn:
    #        l = l.rstrip('\n')
    #        if l[0] == "#":
    #            continue
    #        sl = l.split('\t')
    #        if sl[1] == "!" or sl[3] == "!":
    #            tot_badreads += 1
    #            continue
    #        r1, r2, c1, s1, c2, s2, p1, p2 = parse_hits(sl)
    #        if abs(p1 - p2) < 5000000:
    #            tot_chromreads += 1
    #            if c1 == c2 and s1 == s2:
    #                add_hits(self_hits, c1,s1,p1, c2,s2,p2, winsize, winstep)
    #            elif c1 == c2 and s1 != s2:
    #                add_hits(pair_hits, c1,s1,p1, c2,s2,p2, winsize, winstep)
    #        tot_goodreads += 1
    return(None)

def conv_all(inconns, coord_conv, mummer_query):
    for i in inconns:
        conv_file(i, coord_conv, mummer_query)
    return(None)

def main():
    
    parser = argparse.ArgumentParser("Count up pairing and chromosome self-interactions in Hi-C .pairs files")
    
    parser.add_argument("input", nargs="*", help="One or more .pairs files to use as input (default = stdin).")
    parser.add_argument("-i", "--standard_input", help="Take standard input and other input files (default = False)", action="store_true")
    parser.add_argument("-q", "--mummer_query", nargs=1, help="If you have a mummer .delta file, specify the genotype that was the mummer query. Its coordinates will be converted to the subject's coordinates.")
    parser.add_argument("-m", "--mummer_delta", nargs=1, help="Mummer .delta file to convert read coordinates.")

    args = parser.parse_args()

    inconns = []
    mummer_query = ""
    mummer_query = ""
    if args.standard_input or not args.input:
        inconns.append(sys.stdin)
    if args.input:
        for i in args.input:
            inconns.append(open(i, "r"))
    if args.mummer_query:
        mummer_query = args.mummer_query
    if args.mummer_delta:
        mummer_delta = args.mummer_delta

    coord_conv = get_coord_conv(mummer_delta)
    print(coord_conv)
    
    conv_all(inconns, coord_conv, mummer_query)
    
    for i in inconns:
        i.close()

if __name__ == "__main__":
    main()

#/Users/jbaldwin/Documents/work_stuff/drosophila/homologous_hybrid_mispairing/refs/mummer_script/wx01_clean.fa /Users/jbaldwin/Documents/work_stuff/drosophila/homologous_hybrid_mispairing/refs/mummer_script/w501_clean.fa
#NUCMER
#>2L X 24213528 22038717
#37728 39355 3481 5108 0 0 0
#0
#49709 51223 3596 5110 0 0 0
#0
#157862 161652 13547 17338 6 6 0
#-3494
#0
#158252 161659 5108 8516 5 5 0
#-3104
#0
#166236 168887 12379755 12377104 11 11 0
#0
#166716 169961 10994715 10991473 13 13 0
#2267
#1
#1
#1
#-39
#0
