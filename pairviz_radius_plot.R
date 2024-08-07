#!/usr/bin/env Rscript

library(ggplot2)

main = function() {
    args = commandArgs(trailingOnly=TRUE)
    inpath = args[1]
    gzinpath = gzfile(inpath, "r")
    outpath = args[2]
    data = read.table(gzinpath, sep="\t", header=FALSE)
    colnames(data) = c("Distance", "Variable", "Value")
    plot = ggplot(data=data,
            mapping = aes(x=Distance, y=Value, color=Variable)) +
        geom_smooth() +
        scale_color_discrete(
            breaks=c("pair_count", "pair_prop", "self_count", "uninf_count"),
            labels = c("Paired reads", "Proportion paired", "Self-interaction reads", "Uninformative")
        ) +
        theme_bw()
    res_scale = 100
    png(outpath, width=8*res_scale, height=6*res_scale, res=res_scale)
    print(plot)
    dev.off()
        #geom_line() +
}
    
main()
