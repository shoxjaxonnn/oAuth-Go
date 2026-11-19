package docs

import (
	_ "embed"
	"net/http"
)

//go:embed openapi.json
var openAPISpec []byte

const uiHTML = `<!DOCTYPE html>
<html>
<head>
  <title>oAuth Demo API docs</title>
  <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/swagger-ui/5.17.14/swagger-ui.min.css">
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://cdnjs.cloudflare.com/ajax/libs/swagger-ui/5.17.14/swagger-ui-bundle.min.js"></script>
  <script>
    window.onload = () => SwaggerUIBundle({ url: '/docs/openapi.json', dom_id: '#swagger-ui' });
  </script>
</body>
</html>`

// Mount registers /docs (a Swagger UI page, loaded from a CDN) and
// /docs/openapi.json (the spec itself, embedded in the binary).
func Mount(mux *http.ServeMux) {
	mux.HandleFunc("GET /docs/openapi.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(openAPISpec)
	})
	serveUI := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(uiHTML))
	}
	mux.HandleFunc("GET /docs", serveUI)
	mux.HandleFunc("GET /docs/", serveUI)
}
