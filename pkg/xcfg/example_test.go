// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package xcfg_test

import (
	"fmt"
	"strings"

	"github.com/ctx42/xcfg/pkg/xcfg"
)

func ExampleNew() {
	src := "# Database.\nDB_HOST=localhost\nDB_PORT=5432\n"
	cfg, _ := xcfg.New(strings.NewReader(src))

	host, _ := cfg.CfgKey("DB_HOST")
	port, _ := xcfg.ParseInteger("DB_PORT", cfg["DB_PORT"])
	fmt.Println(host, port)

	_, err := cfg.CfgKey("DB_USER")
	fmt.Println(err)
	// Output:
	// localhost 5432
	// DB_USER: config key missing
}

func ExampleNewFromPath() {
	enc := xcfg.Config{"LOG_LEVEL": "info"}.Encode()
	env := []string{"LOG_LEVEL=debug"}

	cfg, _ := xcfg.NewFromPath(enc, xcfg.WithEnv(env))

	fmt.Println(enc)
	fmt.Println(cfg["LOG_LEVEL"])
	// Output:
	// base64:TE9HX0xFVkVMPSJpbmZvIg
	// debug
}

func ExampleAddAdapter() {
	cfg := xcfg.Config{"DB_HOST": "localhost"}
	db := xcfg.AddAdapter(cfg, "DB")

	host, _ := db.CfgKey("HOST")
	fmt.Println(host)
	// Output:
	// localhost
}
