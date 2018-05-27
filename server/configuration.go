/*
 * Revision History:
 *     Initial: 2018/5/26        ShiChao
 */

package server

// Configuration for a http server.
type Configuration struct {
	Address string
}

// TLSConfiguration is the configuration for a https server.
type TLSConfiguration struct {
	Key  string
	Cert string
}
