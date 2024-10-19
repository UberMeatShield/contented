Ansible code needs to be written to do a lot of this but developing proper ansible without Vagrant & virtualbox is pretty annoying.


Helpful command:
sudo systemctl daemon-reload && sudo systemctl stop contented.service && sleep 1 && sudo systemctl start contented.service && journalctl -u contented.service -f

dnf install nginx
dnf install golang
dnf install nodejs


mkdir -p /etc/pki/nginx/private
chown nginx:nginx
sudo openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout /etc/pki/nginx/private/server.key -out /etc/pki/nginx/server.crt
Install buffalo
Create a user that is NOT the ec2-user who has docker permissions

# docker
dnf install docker
wget https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)
sudo mv docker-compose-$(uname -s)-$(uname -m) /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

Add ec2 user to docker group


    gzip on;
    gzip_types
      application/javascript
      application/json
      application/xhtml+xml
      text/css
      text/javascript
      text/html
      text/plain;

    upstream buffalo_app {
        server 127.0.0.1:3000;
    }

    server {
        listen       443 ssl http2;
        listen       [::]:443 ssl http2;
        server_name  _;
        root         /usr/share/nginx/html;

        ssl_certificate "/etc/pki/nginx/server.crt";
        ssl_certificate_key "/etc/pki/nginx/private/server.key";
        ssl_session_cache shared:SSL:1m;
        ssl_session_timeout  10m;
        ssl_ciphers PROFILE=SYSTEM;
        ssl_prefer_server_ciphers on;

        # Load configuration files for the default server block.
        include /etc/nginx/default.d/*.conf;

        server_tokens off;

        location / {
          proxy_redirect   off;
          proxy_set_header Host              $http_host;
          proxy_set_header X-Real-IP         $remote_addr;
          proxy_set_header X-Forwarded-For   $proxy_add_x_forwarded_for;
          proxy_set_header X-Forwarded-Proto $scheme;

          proxy_pass       http://buffalo_app;
        }

        error_page 404 /404.html;
        location = /404.html {
        }

        error_page 500 502 503 504 /50x.html;
        location = /50x.html {
        }
    }





Ensure the environment file actually gets updated with something resembling valid
