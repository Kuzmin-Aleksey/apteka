function getCookie(e, t = !1) {
    if (!e) return;
    let n = document.cookie.match(new RegExp("(?:^|; )" + e.replace(/([.$?*|{}()\[\]\\\/+^])/g, "\\$1") + "=([^;]*)"));
    if (n) {
        let e = decodeURIComponent(n[1]);
        if (t) try {
            return JSON.parse(e)
        } catch (e) {
        }
        return e
    }
}

function setCookie(e, t, n = {path: "/"}) {
    if (!e) return;
    (n = n || {}).expires instanceof Date && (n.expires = n.expires.toUTCString()), t instanceof Object && (t = JSON.stringify(t));
    let o = encodeURIComponent(e) + "=" + encodeURIComponent(t);
    for (let e in n) {
        o += "; " + e;
        let t = n[e];
        !0 !== t && (o += "=" + t)
    }
    document.cookie = o
}

function deleteCookie(e) {
    setCookie(e, null, {expires: new Date, path: "/"})
}