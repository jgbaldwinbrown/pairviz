#!/usr/bin/env python3

import sys
import plotnine as p9
import statsmodels.api as sm
import seaborn as sns
import scipy.signal as sig
import pandas as pd
import fractions

def coords2num(row, chr_ranks):
    return chr_ranks[row["chrom"]] * 1e12 + int(row["start"])

def loessify_data(data, frac, chr_ranks):
    data["num_pos"] = data.apply(lambda row: coords2num(row, chr_ranks), axis=1)
    data["lowess"] = sm.nonparametric.lowess(data["pair_prop_fpkm"],
        data["num_pos"],
        frac = frac,
        return_sorted = False)
    data["negative_lowess"] = -data["lowess"]

def get_peaks(data, sign):
    prom = 0.02
    if sign == "-":
        peaks = sig.find_peaks(data["negative_lowess"], prominence = prom)[0].tolist()
        peak_vals = data.iloc[peaks]
        peak_vals = peak_vals.assign(direction = "valley")
    else:
        peaks = sig.find_peaks(data["lowess"], prominence = prom)[0].tolist()
        peak_vals = data.iloc[peaks]
        peak_vals = peak_vals.assign(direction = "peak")
    return (peaks, peak_vals)

def print_data(data, outpath):
    data.to_csv(outpath, index=False, compression="gzip")

def print_peaks(peaks, outpath):
    with open(outpath, "w") as outconn:
        for peak in peaks:
            outconn.write(str(peak) + "\n")

def print_peak_vals(peak_vals, outpath):
    peak_vals.to_csv(outpath, index=False)

def get_chr_ranks(rankpath):
    chr_ranks = {}
    with open(rankpath, "r") as inconn:
        for l in inconn:
            sl = l.rstrip('\n').split('\t')
            chr_ranks[sl[0]] = int(sl[1])
    return chr_ranks

def combine_peak_vals(peak_vals, neg_peak_vals):
    return peak_vals.append(neg_peak_vals, ignore_index=True)

def main():
    data = pd.read_csv(sys.argv[1], sep="\t", header=0)
    frac = fractions.Fraction(sys.argv[2])
    chr_ranks = get_chr_ranks(sys.argv[3])
    loessify_data(data, frac, chr_ranks)
    print_data(data, sys.argv[4] + "_data.txt.gz")
    peaks, peak_vals = get_peaks(data, "+")
    neg_peaks, neg_peak_vals = get_peaks(data, "-")
    peak_vals = combine_peak_vals(peak_vals, neg_peak_vals)
    print_peaks(peaks, sys.argv[4] + "_peaks.txt")
    print_peaks(neg_peaks, sys.argv[4] + "_neg_peaks.txt")
    print_peak_vals(peak_vals, sys.argv[4] + "_peak_vals.txt")

if __name__ == "__main__":
    main()

# def main():
#     iris = sns.load_dataset('iris')
#     iris.sort_values(by="sepal_length", inplace=True)
#     # print(iris)
#     x=iris["sepal_length"]
#     y=iris["sepal_width"]
#     y_hat1 = sm.nonparametric.lowess(y, x, return_sorted=False)
#     y_hat2 = sm.nonparametric.lowess(y, x, frac=1/5, return_sorted=False)
#     y_hat3 = sm.nonparametric.lowess(y, x, frac=9/10, return_sorted=False)
#     iris["y_hat1"] = y_hat1
#     iris["y_hat2"] = y_hat2
#     iris["y_hat3"] = y_hat3
#     aplot = (
#         p9.ggplot(data = iris,
#             mapping = p9.aes(x = "sepal_length", y = "sepal_width")
#         ) +
#         p9.geom_point() +
#         p9.geom_line(mapping = p9.aes(x="sepal_length", y="y_hat1"), color="red") +
#         p9.geom_line(mapping = p9.aes(x="sepal_length", y="y_hat2"), color="blue") +
#         p9.geom_line(mapping = p9.aes(x="sepal_length", y="y_hat3"), color="green")
#     )
#     # aplot.save("temp.pdf", height=3, width=4)
#     peaks2 = sig.find_peaks(iris["y_hat2"])[0].tolist()
#     peaks2_1 = sig.find_peaks(iris["y_hat2"], prominence=0.1)[0].tolist()
#     peaks3 = sig.find_peaks(iris["y_hat3"])
#     # print(peaks2)
#     # print(peaks2_1)
#     peaks3_1 = sig.find_peaks(iris["y_hat3"], prominence = 0.1)[0]
#     proms3 = sig.peak_prominences(iris["y_hat3"], peaks3[0].tolist())
#     peaks2_vals = iris.iloc[peaks2]
#     peaks2_1_vals = iris.iloc[peaks2_1]
#     print(peaks2_vals)
#     print(peaks2_1_vals)
#     peaks3_vals = iris.iloc[peaks3_1]
#     # print(peaks3_vals)
