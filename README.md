# Deezer Badge - Show your last played song

This golang Project allows you to create a badge to show off your great music taste to the world.

![Deezer Now Playing](https://incredible.software/test/badge/07371d90-f3ce-4352-b50a-93b55e3102e9)

Be aware that this server stores your userId + an offline access token which does not expire.
The access token also allows the server owner to query the /me endpoint end exfiltrate all data visible there.
So pleas set up your own server or run one for your friends, don't use strangers server.

Leave me a star if you find this project helpful :)

# Setup

You can run this app easy with docker compose:

```bash
git clone https://github.com/cli-ish/deezer-badge.git
cd deezer-badge/
docker compose build
cp .env.sample .env
# Now create a deezer app at https://developers.deezer.com/myapps
# Fill out the .env file with the APP_ID, APP_SECRET and RETURN_URL

docker compose up -d
# Its also highly recommended to use a ssl proxy such as caddy or nginx to upgrade the connection to ssl.
# If you want to configure it to an external http server just delete the "127.0.0.1:" from the docker-compose.yml file.
```

A nginx location for this project could look like this:

```
location /test/ {
    expires off;
    proxy_pass http://127.0.0.1:6969/;
    break;
}
```
