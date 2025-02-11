package main

import (
	"fmt"

	"github.com/cxbelka/winter_2025/internal/app"
)

func main() {
	// отладка graceful shutdown
	// go func() {
	// 	time.Sleep(5 * time.Second)
	// 	syscall.Kill(os.Getpid(), syscall.SIGINT)
	// }()

	application, err := app.New()
	if err != nil {
		panic(fmt.Errorf("app init failed: %w", err))
	}
	if err := application.Run(); err != nil {
		panic(fmt.Errorf("app Run failed: %w", err))
	}
}
