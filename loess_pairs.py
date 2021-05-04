#!/usr/bin/env python3

import plotnine as p9
import statsmodels.api as sm
import seaborn as sns
import scipy.signal as sig
import pandas as pd

def loessify_data(data, frac):
    data["lowess"] = sm.nonparametric.lowess(data["pair_prop_fpkm"]

def main():
    data = pd.read_csv(sys.argv[1], sep="\t", header=1)
    frac = float(sys.argv[2])
    loessify_data(data, frac)
    peaks = get_peaks(data)
    print_peaks(peaks, sys.argv[3])

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
