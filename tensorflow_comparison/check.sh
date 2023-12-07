#!/bin/bash
set -e

mypy --ignore-missing-imports tensorflow_predict.py
