#!/usr/bin/env python3

def add_hit(hits, c, s, p, winsize, winstep):
    if c not in hits:
        hits[c] = {}
    base_window = (p // winsize) * winsize
    window_starts = range(base_window, base_window + winsize, winstep)
    windows = [(x, x + winsize) for x in window_starts]
    for w in windows:
        if w not in hits[c]:
            hits[c][w] = 0
        hits[c][w] += 1

def add_hits(hits, c1, s1, p1, c2, s2, p2, winsize, winstep):
    add_hit(hits, c1, s1, p1, winsize, winstep)
    add_hit(hits, c2, s2, p2, winsize, winstep)

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
    return(r1, r2, c1, s1, c2, s2, p1, p2)

def print_hits(hits, alt_hits, hit_type, alt_hit_type, tot_goodreads, tot_badreads, tot_chromreads, winsize, winstep):
    #print(hits)
    #print(alt_hits)
    print("chrom\tstart\tend\thit_type\talt_hit_type\thits\talt_hits\tpair_prop\talt_prop\tpair_totprop\tpair_totgoodprop\tpair_totcloseprop\twinsize\twinstep")
    for chrom in sorted(hits):
        for pos in sorted(hits[chrom]):
            #print(hits[chrom][pos])
            count = hits[chrom][pos]
            try:
                altcount = alt_hits[chrom][pos]
                pairprop = count / (count + altcount)
                altprop = altcount / (count + altcount)
            except KeyError:
                altcount = 0
                pairprop = 1
                altprop = 0
            
            print( "\t".join( map( str, (
                            chrom,
                            pos[0],
                            pos[1],
                            hit_type,
                            alt_hit_type,
                            count,
                            "%.8g" % (altcount),
                            "%.8g" % (pairprop),
                            "%.8g" % (altprop),
                            "%.8g" % (count / (tot_goodreads+tot_badreads)),
                            "%.8g" % (count / (tot_goodreads)),
                            "%.8g" % (count / (tot_chromreads)),
                            winsize,
                            winstep,
            ))))

if __name__ == "__main__":
    import sys

    winsize = int(sys.argv[1])
    winstep = int(sys.argv[2])

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
        r1, r2, c1, s1, c2, s2, p1, p2 = parse_hits(sl)
        if abs(p1 - p2) < 5000000:
            tot_chromreads += 1
            if c1 == c2 and s1 == s2:
                add_hits(self_hits, c1,s1,p1, c2,s2,p2, winsize, winstep)
            elif c1 == c2 and s1 != s2:
                add_hits(pair_hits, c1,s1,p1, c2,s2,p2, winsize, winstep)
        tot_goodreads += 1

    print_hits(pair_hits, self_hits, "paired", "self", tot_goodreads, tot_badreads, tot_chromreads, winsize, winstep)
