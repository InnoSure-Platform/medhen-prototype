package main

import "github.com/InnoSure-Platform/pc-platform/internal/svcboot"

func main() { svcboot.Start("pc-reporting-svc", ":8111", "reporting") }
