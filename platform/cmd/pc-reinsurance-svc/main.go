package main

import "github.com/InnoSure-Platform/pc-platform/internal/svcboot"

func main() { svcboot.Start("pc-reinsurance-svc", ":8112", "reinsurance") }
