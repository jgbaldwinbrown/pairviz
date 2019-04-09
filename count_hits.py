#!/usr/bin/env python3

import sys

self_hits = {}
pair_hits = {}
tot_badreads = 0
tot_goodreads = 0
tot_chromreads = 0
for l in sys.stdin:
    l = l.rstrip('\n')
    if l[0] == "#":
        continue
    sl = l.split('\t')
    if sl[1] == "!" or sl[3] == "!":
        tot_badreads += 1
        continue
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
    if abs(p1 - p2) < 5000000:
        tot_chromreads += 1
        if c1 == c2 and s1 == s2:
            name = c1
            if name not in self_hits:
                self_hits[name] = 0
            self_hits[name] += 1
        elif c1 == c2 and s1 != s2:
            name = c1
            if name not in pair_hits:
                pair_hits[name] = 0
            pair_hits[name] += 1
    tot_goodreads += 1

for k,v in self_hits.items():
    print("self\t" + k + "\t" + str(v))

for k,v in pair_hits.items():
    print("pair\t" + k + "\t" + str(v))

for k,v in pair_hits.items():
    print("pair_proportion\t" + k + "\t" + str(v / (v+self_hits[k])))

for k,v in pair_hits.items():
    print("pair_proportion_of_total_good\t" + k + "\t" + str(v / (tot_goodreads)))

for k,v in pair_hits.items():
    print("pair_proportion_of_total\t" + k + "\t" + str(v / (tot_goodreads + tot_badreads)))

for k,v in pair_hits.items():
    print("pair_proportion_of_total_close_range\t" + k + "\t" + str(v / (tot_chromreads)))

print("total reads: " + str(tot_goodreads + tot_badreads))
print("total good reads: " + str(tot_goodreads))
print("total bad reads: " + str(tot_badreads))
print("total close range reads (< 5Mb): " + str(tot_chromreads))
