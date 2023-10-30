#!/usr/bin/Rscript

library(ggplot2)
library(data.table)
library(reshape2)

plotit = function(path, out, name, mindist, maxdist) {
	d = as.data.frame(fread(path, sep = "\t"))
	colnames(d) = c("Distance", "Pair", "Self", "Trans")
	d = d[d$Distance < maxdist & d.Distance >= mindist,]
	m = melt(d, id = c("Distance"))
	p = ggplot(m, aes(Distance, value, color = factor(variable))) + 
		geom_line() + 
		scale_y_log10() +
		ggtitle(name)
	res = 300
	png(out, width = 4*res, height = 3*res, res=res)
	print(p)
	dev.off()
}

main = function() {
	args = commandArgs(trailingOnly = TRUE)
	if (len(args) != 2) {
		exit("wrong number of arguments")
	}

	inpath = args[1]
	outpath = args[2]
	name = args[3]
	mindist = as.numeric(args[4])
	maxdist = as.numeric(args[5])
	plotit(inpath, outpath)
}

main()

