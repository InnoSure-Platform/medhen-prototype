// Command medhen-api — monolith fallback (same as pc-gateway, single process).
package main

import "github.com/InnoSure-Platform/pc-platform/internal/svcboot"

func main() { svcboot.Main("medhen-api", ":8080", "all") }
