/*
 * Revision History:
 *     Initial: 2018/05/28        ShiChao
 */

package server


const (
	// charset
	charsetUTF8 = "charset=UTF-8"

	// HTTP methods
	PUT     = "PUT"
	GET     = "GET"
	POST    = "POST"
	HEAD    = "HEAD"
	PATCH   = "PATCH"
	DELETE  = "DELETE"
	OPTIONS = "OPTIONS"

	// MIME
	MIMEApplicationJSON            = "application/json"
	MIMEApplicationJSONCharsetUTF8 = MIMEApplicationJSON + "; " + charsetUTF8
	MIMEMultipartForm              = "multipart/form-data"

	// Headers
	HeaderOrigin        = "Origin"
	HeaderAccept        = "Accept"
	HeaderVary          = "Vary"
	HeaderCookie        = "Cookie"
	HeaderSetCookie     = "Set-Cookie"
	HeaderUpgrade       = "Upgrade"
	HeaderContentType   = "Content-Type"
	HeaderLocation      = "Location"
	HeaderAuthorization = "Authorization"

	// Access control
	HeaderAccessControlRequestMethod    = "Access-Control-Request-Method"
	HeaderAccessControlAllowMethods     = "Access-Control-Allow-Methods"
	HeaderAccessControlRequestHeaders   = "Access-Control-Request-Headers"
	HeaderAccessControlAllowOrigin      = "Access-Control-Allow-Origin"
	HeaderAccessControlAllowHeaders     = "Access-Control-Allow-Headers"
	HeaderAccessControlExposeHeaders    = "Access-Control-Expose-Headers"
	HeaderAccessControlAllowCredentials = "Access-Control-Allow-Credentials"
	HeaderAccessControlMaxAge           = "Access-Control-Max-Age"
)