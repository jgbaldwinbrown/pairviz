#!/bin/bash
set -e

go run make_pairtools.go

rsync -avP \
	./ \
	u6012238@lonepeak.chpc.utah.edu:/uufs/chpc.utah.edu/common/home/shapiro-group3/jim/new/fly/hic2 \
	--files-from <(echo Makefile;
		find runscripts -type f)
