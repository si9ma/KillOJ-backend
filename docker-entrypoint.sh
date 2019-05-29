#!/usr/bin/env bash

[ "$1" != "backend" ] && $@ && exit 0

./backend ${@:2}