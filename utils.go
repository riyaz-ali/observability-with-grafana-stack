package main

import "log"

func Must[T any](obj T, err error) T {
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	return obj
}
