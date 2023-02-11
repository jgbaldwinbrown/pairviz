#!/bin/bash
set -e

(cd cmd && (
	go build downsample_pairviz_uniques.go
	go build downsample_beds.go
	go build downsample_bams_uniques.go
	go build make_downsample_config.go
))

cp -p cmd/make_downsample_config ~/mybin/
cp -p cmd/downsample_bams_uniques ~/mybin/
cp -p cmd/downsample_pairviz_uniques ~/mybin/
cp -p cmd/downsample_beds ~/mybin/
