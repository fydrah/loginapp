/*
Copyright 2018 fydrah

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Some code comes from @ericchiang (Dex - CoreOS)
package main

import (
	"html/template"
	"log"
	"net/http"
)

func renderTemplate(w http.ResponseWriter, tmpl *template.Template, data interface{}) {
	err := tmpl.Execute(w, data)
	if err == nil {
		return
	}

	switch err := err.(type) {
	case *template.Error:
		// An ExecError guarantees that Execute has not written to the underlying reader.
		log.Printf("Error rendering template %s: %s", tmpl.Name(), err)

		// TODO(ericchiang): replace with better internal server error.
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	default:
		// An error with the underlying write, such as the connection being
		// dropped. Ignore for now.
	}
}

var indexTmpl = template.Must(template.New("index.html").Parse(`<html>
  <head>
  <link rel="stylesheet" type="text/css" href="assets/css/form-style.min.css">
  <script type="text/javascript">function adjust_textarea(t){t.style.height="20px",t.style.height=t.scrollHeight+"px"}</script>
  </head>
  <body>
      <form class="form-style-7" action="/login" method="post">
        <h1>Kubernetes Loginapp</h1>
        <ul>
          <li>
            <label for="access">CLI access</label>
            <input name="access" type="submit" value="CLI" />
            <span>Login and retrieve your kubeconfig for CLI authentication</span>
          </li>
        </ul>
      </form>
  </body>
</html>`))

var tokenTmpl = template.Must(template.New("token.html").Parse(`<html>
<head>
   <link rel="stylesheet" type="text/css" href="assets/css/code-box-copy.min.css">
   <link rel="stylesheet" type="text/css" href="assets/css/prism.min.css">
   <link rel="stylesheet" type="text/css" href="assets/css/form-style.min.css">
</head>
<body>
   <div class="form-style-7">
      <h1>Kubernetes Loginapp</h1>
      <form class="form-style-7" action="/" method="get">
         <ul>
            <li>
               <label>Copy this in your ~/.kube/config file</label>
               <div class="code-box-copy">
                  <button class="code-box-copy__btn" title="Copy" type="button" data-clipboard-target="#kubeconfig"></button>
                  <pre><code class="language-yaml" id="kubeconfig">- name: {{ .Claims.name }}
  user:
    auth-provider:
      config:
        client-id: {{ .ClientID }}
        client-secret: {{ .ClientSecret }}
        id-token: {{ .IDToken }}
        idp-issuer-url: {{ .Claims.iss }}
        refresh-token: {{ .RefreshToken }}
      name: oidc</code></pre>
               </div>
            </li>
            <li>
               <label>Return to login page</label>
               <input type="submit" value="Home" />
            </li>
         </ul>
      </form>
   </div>
   <script src="assets/js/clipboard.min.js"></script>
   <script src="assets/js/jquery.min.js"></script>
   <script src="assets/js/prism.min.js"></script>
   <script src="assets/js/code-box-copy.min.js"></script>
   <script>
      $('.code-box-copy').codeBoxCopy();
   </script>
</body>
</html>
`))
