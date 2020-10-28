# Instagram API access token

[Follow tutorial](https://developers.facebook.com/docs/instagram-basic-display-api/getting-started)
to configure Facebook app to access Instagram API.
You will need to grant the app access to your Instagram profile —
open a URL in browser with your app credentials.

```sh
app_id=123456789101112
app_secret=example81b0eb4b009432bb62e55c9cd7
redirect_uri=https://localhost:8000/

open https://api.instagram.com/oauth/authorize?client_id=${app_id}&redirect_uri=${redirect_uri}&scope=user_profile,user_media&response_type=code
```

Copy the code from URL `?code=...` you were redirected to after granting access.
Note, don't copy `#_` part.

Get a short lived token and optionally a long lived token.

```sh
code=ABCD9...XRZ-v

short_lived_token=$(curl -s -X POST https://api.instagram.com/oauth/access_token \
    -F client_id=${app_id} \
    -F client_secret=${app_secret} \
    -F grant_type=authorization_code \
    -F redirect_uri=${redirect_uri} \
    -F code=${code} | jq '.access_token' \
)
echo ${short_lived_token}

long_lived_token=$(curl -s "https://graph.instagram.com/access_token?grant_type=ig_exchange_token&client_secret=${app_secret}&access_token=${short_lived_token}" \
    | jq '.access_token' \
)
echo ${long_lived_token}
```
