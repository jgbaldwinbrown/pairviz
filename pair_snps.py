#!/usr/bin/env python3

import sys
import argparse

def get_snps(mummer_snps, mummer_reference, mummer_query):
    geno1 = mummer_reference
    geno2 = mummer_query
    snps = {}
    snps[geno1] = {}
    snps[geno2] = {}
    #1,4,12,13
    for i,l in enumerate(open(mummer_snps, "r")):
        l=l.rstrip('\n')
        sl = l.split()
        if i<5:
            continue
        p1, p2, c1, c2 = (int(sl[0]), int(sl[3]), sl[13], sl[14])
        if c1 not in snps[geno1]:
            snps[geno1][c1] = {}
        if c2 not in snps[geno2]:
            snps[geno2][c2] = {}
        snps[geno1][c1][p1] = 1
        snps[geno2][c2][p2] = 1
    return(snps)


def parse_hits(sl):
    #collect info about read alignments
    r1 = sl[1].split('_')
    r2 = sl[3].split('_')
    #get (c)hromosome and (s)pecies for each of the 2 reads
    c1 = r1[0]
    s1 = r1[1]
    c2 = r2[0]
    s2 = r2[1]
    p1 = int(sl[2])
    p2 = int(sl[4])
    dir1 = sl[5]
    dir2 = sl[6]
    return(r1, r2, c1, s1, c2, s2, p1, p2, dir1, dir2)

def filter(c1, s1, c2, s2, p1, p2, dir1, dir2, readlen, snps, snpsreq):
    r1snps = 0
    r2snps = 0
    r1end = p1 + readlen if dir1 == "+" else p1 - readlen
    r2end = p2 + readlen if dir2 == "+" else p2 - readlen
    r1dir = 1 if dir1 == "+" else -1
    r2dir = 2 if dir2 == "+" else -1
    for r1pos, r2pos in zip(range(p1,r1end,r1dir), range(p2,r2end,r2dir)):
        #sys.stdout.write("\t" + str(r1pos) + ":" + str(r2pos))
        #try:
        #    sys.stdout.write("\tsnps1:"+str(snps[s1][c1][r1pos]))
        #except:
        #    pass
        #try:
        #    sys.stdout.write("\tsnps2:"+str(snps[s2][c2][r2pos]))
        #except:
        #    pass
        try:
            if r1pos in snps[s1][c1]:
                r1snps += 1
            if r2pos in snps[s2][c2]:
                r2snps += 1
        except KeyError:
            pass
    #sys.stdout.write("\n")
    return(r1snps >= snpsreq and r2snps >= snpsreq)

def filter_file(inconn, snps, readlen, snpsreq):
    #print(snps)
    for l in inconn:
        l = l.rstrip('\n')
        if l[0] == "#":
            print(l)
            continue
        sl = l.split('\t')
        if sl[1] == "!" or sl[3] == "!":
            continue
        r1, r2, c1, s1, c2, s2, p1, p2, dir1, dir2 = parse_hits(sl)
        #print(sl)
        #print("c1, s1, c2, s2, p1, p2, dir1, dir2, readlen")
        #print(c1, s1, c2, s2, p1, p2, dir1, dir2, readlen)
        #print(filter(c1, s1, c2, s2, p1, p2, dir1, dir2, readlen, snps))
        if filter(c1, s1, c2, s2, p1, p2, dir1, dir2, readlen, snps, snpsreq):
            print(l)
    return(None)

def filter_all(inconns, snps, readlen, snpsreq):
    for i in inconns:
        filter_file(i, snps, readlen, snpsreq)
    return(None)

def main():
    
    parser = argparse.ArgumentParser("Count up pairing and chromosome self-interactions in Hi-C .pairs files")
    
    parser.add_argument("input", nargs="*", help="One or more .pairs files to use as input (default = stdin).")
    parser.add_argument("-i", "--standard_input", help="Take standard input and other input files (default = False)", action="store_true")
    parser.add_argument("-m", "--mummer_snps", required=True, help="Mummer .snps file to filter on.")
    parser.add_argument("-q", "--mummer_query", required=True, help="Specify the genotype that was the mummer query in the mummer .snps file.")
    parser.add_argument("-r", "--mummer_reference", required=True, help="Specify the genotype that was the mummer reference in the mummer .snps file.")
    parser.add_argument("-l", "--read_length", help="Specify the length of the sequencing reads. Default = 125bp.")
    parser.add_argument("-s", "--snps_required", help="Number of SNPs required to consider a read discriminable. Default = 1.")

    args = parser.parse_args()

    inconns = []
    mummer_query = ""
    mummer_reference = ""
    mummer_snps = ""
    readlen = 125
    snpsreq = 1
    if args.standard_input or not args.input:
        inconns.append(sys.stdin)
    if args.input:
        for i in args.input:
            inconns.append(open(i, "r"))
    if args.mummer_query:
        mummer_query = args.mummer_query
    if args.mummer_reference:
        mummer_reference = args.mummer_reference
    if args.mummer_snps:
        mummer_snps = args.mummer_snps
    if args.read_length:
        readlen = int(args.read_length)
    if args.snps_required:
        snpsreq = int(args.snps_required)

    snps = get_snps(mummer_snps, mummer_reference, mummer_query)
    
    filter_all(inconns, snps, readlen, snpsreq)
    
    for i in inconns:
        i.close()

if __name__ == "__main__":
    main()

#
#/Users/jbaldwin/Documents/work_stuff/drosophila/homologous_hybrid_mispairing/refs/mummer_script/iso1_clean.fa /Users/jbaldwin/Documents/work_stuff/drosophila/homologous_hybrid_mispairing/refs/mummer_script/w501_clean.fa
#NUCMER

#    [P1]  [SUB]  [P2]      |   [BUFF]   [DIST]  |  [LEN R]  [LEN Q]  | [FRM]  [TAGS]
#========================================================================================
#   10376   A T   79576     |        6    10376  | 23513712 25159701  |  1  1  2L	2L
#   10384   T G   79584     |        8    10384  | 23513712 25159701  |  1  1  2L	2L
#   10402   A C   79602     |        1    10402  | 23513712 25159701  |  1  1  2L	2L
#   10403   C A   79603     |        1    10403  | 23513712 25159701  |  1  1  2L	2L
#   10406   C G   79606     |        3    10406  | 23513712 25159701  |  1  1  2L	2L
#
#
#
#D00550:549:CD5KUANXX:5:1101:1128:31971  2L_W501 15145684        2L_W501 15146216        +       -       UU
#D00550:549:CD5KUANXX:5:1101:1128:32917  3L_W501 23115097        3L_W501 23115574        +       -       UU
