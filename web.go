package main

import "embed"

//go:embed all:dist
var web embed.FS
