package cruntime

import _ "embed"

//go:embed parson.h.txt
var ParsonH string

//go:embed parson.c.txt
var ParsonC string

//go:embed kodae_bootstrap.txt
var BootstrapC string

//go:embed ws_client.c.txt
var WsClientC string
