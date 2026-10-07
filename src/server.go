package main

import (
	"fmt"
	"log"
	"net/http"
)

func homeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	htmlString := `
	    <!DOCTYPE html>
    <html>
      <head>
        <title>SaasySquad</title>
        <style>
          body {
            font-family: sans-serif;
            display: flex;
            flex-direction: column;
            align-items: center;
            justify-content: center;
            min-height: 100vh;
            margin: 0;
            background-color: #f9fafb;
          }
          h1 { color: #333; }
          .gif-container { margin-top: 20px; }
          img { max-width: 300px; margin: 0 10px; border-radius: 8px; }
          
          /* Login Button Styles */
          .login-btn {
            margin-top: 20px;
            padding: 12px 24px;
            background-color: #4285F4;
            color: white;
            text-decoration: none;
            border-radius: 4px;
            font-weight: bold;
            transition: background-color 0.2s;
            display: inline-flex;
            align-items: center;
            gap: 10px;
          }
          .login-btn:hover {
            background-color: #357ae8;
          }
          .token-display {
            margin-top: 20px;
            padding: 15px;
            background: #eef2ff;
            border: 1px solid #c7d2fe;
            border-radius: 8px;
            max-width: 80%;
            word-break: break-all;
            font-family: monospace;
          }
          .hidden { display: none; }
        </style>
      </head>
      <body>
        <h1>You've Reached The Root Of SaasySquad :O</h1>
        
        <a href="https://jasonlihav1234.github.io/sassysquad-backend/api-docs/" target="_blank" rel="noopener noreferrer">Our Swagger API</a>

        <div id="auth-section">
          <a href="/auth/google/login" class="login-btn">
            <svg width="18" height="18" viewBox="0 0 18 18" xmlns="http://www.w3.org/2000/svg"><path d="M17.64 9.2c0-.637-.057-1.251-.164-1.84H9v3.481h4.844c-.209 1.125-.843 2.078-1.796 2.717v2.258h2.908c1.702-1.567 2.684-3.874 2.684-6.615z" fill="#4285F4"/><path d="M9 18c2.43 0 4.467-.806 5.956-2.184l-2.908-2.259c-.806.54-1.837.86-3.048.86-2.344 0-4.328-1.584-5.036-3.711H.957v2.332A8.997 8.997 0 009 18z" fill="#34A853"/><path d="M3.964 10.712c-.18-.54-.282-1.117-.282-1.712s.102-1.173.282-1.712V4.956H.957a8.991 8.991 0 000 8.088l3.007-2.332z" fill="#FBBC05"/><path d="M9 3.58c1.321 0 2.508.454 3.44 1.345l2.582-2.58C13.463.891 11.426 0 9 0A8.997 8.997 0 00.957 4.956L3.964 7.29C4.672 3.163 6.656 3.58 9 3.58z" fill="#EA4335"/></svg>
            Sign in with Google
          </a>
        </div>

        <div id="token-section" class="hidden">
          <p style="color: green; font-weight: bold;">✓ Authenticated successfully!</p>
          <div class="token-display" id="token-data"></div>
        </div>

        <div class="gif-container">
          <img src="https://media1.tenor.com/m/JHuU14ekU3EAAAAd/ishowspeed-deglove.gif" alt="SaasySquad GIF 1" />
          <img src="https://i.makeagif.com/media/11-08-2024/HSMtFe.gif" alt="SaasySquad GIF 2" />
        </div>

        <script>
          // Automatically check for a token in the URL hash (from your callback redirect)
          window.addEventListener('load', () => {
            const hash = window.location.hash.substring(1);
            const params = new URLSearchParams(hash);
            const accessToken = params.get("access_token");

            if (accessToken) {
              // Hide login, show token
              document.getElementById('auth-section').classList.add('hidden');
              const tokenSection = document.getElementById('token-section');
              tokenSection.classList.remove('hidden');
              document.getElementById('token-data').innerText = "Access Token: " + accessToken;

              // Clear the hash from URL for cleanliness
              window.history.replaceState(null, "", "/");
            }
          });
        </script>
      </body>
    </html>
	`

	w.Write([]byte(htmlString))
}

func main() {
	http.HandleFunc("GET /", homeHandler)

	fmt.Println("Server running on http://localhost:7000")

	log.Fatal(http.ListenAndServe(":7000", nil))
}
