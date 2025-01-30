
export function fetch_params_web({method, json, form, headers}) {
  // Axios my ass
  let body;
  if (!headers)
    headers = {};

  if (json) {
    headers['Content-Type'] ='application/json';
    body = JSON.stringify(json)
  } else if (form) {
//    headers['Content-Type'] ='application/json';
    body = new FormData();
    for (let [k,v] of Object.entries(form))
      body.append(k, v);
  }

  return {
    method,
    cache: 'no-cache',
    credentials: 'same-origin',
    headers: {
      'X-CSRF-Token': getCookie('dave-auth-token-csrf'),
      ...headers
    },
    body
  };
}

function getCookie(name) {
  const cookieArr = document.cookie.split(';');

  for (let i = 0; i < cookieArr.length; i++) {
    const cookiePair = cookieArr[i].split('=');

    if (name === cookiePair[0].trim()) {
      return decodeURIComponent(cookiePair[1]);
    }
  }

  return null;
}
