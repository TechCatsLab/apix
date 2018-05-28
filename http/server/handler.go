/*
 * Revision History:
 *     Initial: 2018/5/26        ShiChao
 */

package server

// HandlerFunc defines a handler function to handle http request.
type HandlerFunc func(*Context) error
