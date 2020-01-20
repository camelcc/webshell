WebShell
=======

Demo for web shell. A web page with shell capability connecting to backend server.

Test
-----

Run `docker-compose up` under current direcotry for local testing purpose. Visit localhost and using admin/admin for testing purpose.

**server**

Server is a go process serve the `sh` func based on `pty`, route all io to incoming websocket.

**client**

Client is a react SPA using antd design library.

**Authentication**

Web session been authenticated through cookie. Username and password is hardcoded in go process.

**Proxy**

Nginx serve as a reverse proxy to host these two services under the same domain to solve the CORS issue.

Deploy
------

**Go**

Authentication is been hardcoded and need to integrated into real backend system such as redis/memcached for web session, and real db for password verification.

Using `go build` to build the binary and serve it.

**Web**

Using `npm buld` and host the output file under the same domain as backend server.

**Reverse Proxy**

Refer `nginx/nginx.conf` for reference. Note if you want to deploy behind the ssl cerficiation, need to tweak web url from ws to wss, also need to enforce the cookie settings as secure cookie.
