#  OAuth Mini System – Curl Commands

## 1. Start Authorization Flow (Client → Auth Server redirect)
curl -v http://localhost:9000/authorize


## 2. Follow Redirect Automatically
curl -L http://localhost:9000/authorize


## 3. Simulate Callback (Client endpoint)
# Replace code if your auth server generates a different one
curl "http://localhost:9000/callback?code=xyz_server_code_123&state=bik123"


## 4. Call Token Endpoint Directly (Auth Server)
curl -v -X POST http://localhost:8080/token 
  -H "Content-Type: application/x-www-form-urlencoded" 
  -d "grant_type=authorization_code" 
  -d "code=xyz_server_code_123" 
  -d "redirect_uri=http://localhost:9000/callback" 
  -d "client_id=BIK123" 
  -d "client_secret=B1K2SH12D"


## 5. Test Wrong Method (Debug – should fail)
curl -v http://localhost:8080/token


## 6. Access Protected Resource (Direct to Auth Server)
curl -v -H "Authorization: Bearer very-unsafe-access-token" 
  http://localhost:8080/resource


## 7. Access Protected Resource via Client
curl -v http://localhost:9000/fetch-protected-resource


## 🔥 Debug Tip
# Add -v to any curl to see:
# - HTTP method
# - headers
# - redirects
# - full response