#!/bin/bash
set -e

(cd cmd && (
	go build downsample_pairviz_uniques.go
))

cp -p cmd/downsample_pairviz_uniques ~/mybin/
