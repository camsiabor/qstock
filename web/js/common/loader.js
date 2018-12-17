
const QLoader = {};

QLoader.debug = false;

QLoader.injectHead = (function (i, n, j, e, c, t, s) {
    t = n.createElement(j);
    s = n.getElementsByTagName(j)[0];
    t.appendChild(n.createTextNode(e.text));
    t.onload = c(e);
    s ? s.parentNode.insertBefore(t, s) : n.head.appendChild(t)
}); // eslint-disable-line


QLoader.fetch = function (inputs, promise) {
    if (!arguments.length)
        return Promise.reject(new ReferenceError("Failed to execute 'fetchInject': 1 argument required but only 0 present."));
    if (arguments[0] && arguments[0].constructor !== Array)
        return Promise.reject(new TypeError("Failed to execute 'fetchInject': argument 1 must be of type 'Array'."));
    if (arguments[1] && arguments[1].constructor !== Promise)
        return Promise.reject(new TypeError("Failed to execute 'fetchInject': argument 2 must be of type 'Promise'."));

    const resources = [];
    const deferreds = promise ? [].concat(promise) : []
    const thenables = [];

    inputs.forEach(input => deferreds.push(
        window.fetch(input).then(res => {
            return [res.clone().text(), res.blob()]
        }).then(promises => {
            return Promise.all(promises).then(resolved => {
                resources.push({text: resolved[0], blob: resolved[1]})
            })
        })
    ));

    return Promise.all(deferreds).then(() => {
        resources.forEach(resource => {
            thenables.push({
                then: resolve => {
                    resource.blob.type.includes('text/css')
                        ? QLoader.injectHead(window, document, 'style', resource, resolve)
                        : QLoader.injectHead(window, document, 'script', resource, resolve)
                }
            })
        });
        return Promise.all(thenables)
    })
};

QLoader.fetch_html_fragment_and_js = function(selector, suffix, urls) {
    return QLoader.fetch_html_fragment(selector, suffix).then(function () {
        return QLoader.fetch_with_suffix( urls, suffix);
    });
};

QLoader.fetch_html_fragment = function(selector, suffix) {

    let doms = document.querySelectorAll(selector);
    if (!doms || !doms.length) {
        return Promise.resolve(false);
    }

    let promises = [];
    for (let i = 0, n = doms.length; i < n; i++) {
        let dom = doms[i];
        let url = dom.getAttribute("include");
        if (!url) {
            continue;
        }
        let dosuffix = dom.className.indexOf("nosuffix") < 0;
        if (dosuffix && suffix) {
            if (url.indexOf("?") >= 0) {
                url = url + suffix;
            } else {
                url = url + "?" + suffix;
            }
        }

        let obj = {
            url : url,
            dom : dom
        };
        promises[i] = window.fetch(url).then(function (resp) {
            return resp.text().then(function (content) {
                this.dom.removeAttribute("include");
                this.dom.setAttribute("included", this.url);
                if (content) {
                    this.dom.innerHTML = content;
                }
            }.bind(this));
        }.bind(obj));
    }
    if (promises.length === 0) {
       return Promise.resolve(false);
    }

    return Promise.all(promises).then(function () {
        // return QLoader.fetch_html_fragment(selector, suffix);
    });

};

/**
 * opts.objs {
 *     url
 *     callback
 * }
 * opts.suffix
 * @param opts
 */
QLoader.fetch_and_do = function(opts) {
    let objs = opts.objs;
    for (let i = 0; i < objs.length; i++) {
        let obj = objs[i];
        if (typeof obj === 'string') {
            objs[i] = { url : obj };
        }
    }
    if (typeof opts.index === 'undefined') {
        opts.index = 0;
    }

    if (opts.index >= opts.objs.length) {
        return true;
    }

    let target = opts.objs[opts.index];
    if (opts.suffix) {
        if (target.url.indexOf('?') >= 0) {
            target.url = target.url + opts.suffix;
        } else {
            target.url = target.url + "?" + opts.suffix;
        }
    }
    return window.fetch(target.url).then(function (resp) {
        if (target.callback) {
            target.callback.call(this, resp);
        }
        opts.index = opts.index + 1;
        return QLoader.fetch_and_do(opts);
    });
};

QLoader.fetch_with_suffix = function (urls, suffix) {

    if (typeof urls === 'string') { urls = [urls]; }
    if (!(urls instanceof Array)) {
       return Promise.resolve(false);
    }
    if (suffix) {
        for (let i = 0; i < urls.length; i++) {
            let url = urls[i];
            if (url.indexOf('?') < 0) {
                url = url + "?" + suffix;
            } else {
                url = url + "&" + suffix;
            }
            urls[i] = url;
        }
    }
    let url = urls[0];
    if (!url) {
        return Promise.resolve(false);
    }

    for(let i = 0, n = urls.length; i < n; i++) {
        let url = urls[i];
    }

    let promise = new Promise(function (resolve, reject) {
        let dom = document.createElement('script');
        dom.setAttribute("type", "text/javascript");
        dom.setAttribute("src", url);
        dom.onload = dom.onreadystatechange = function () {
            if (QLoader.debug) {
                console.log("[fetch]", dom);
            }
            resolve(this);
        }.bind(urls.slice(1));
        document.body.appendChild(dom);
    });
    promise.then(function (urls) {
        if (!urls || urls.length === 0) {
            return true;
        }
        return QLoader.fetch_with_suffix(urls);
    });
    return promise;
};

QLoader.fetch_if = function (urls, suffix, opt) {
    opt = opt || {};
    if (opt.url) {
        if (window.location.href.indexOf(opt.url) >= 0) {
            return QLoader.fetch_with_suffix(urls, suffix);
        }
    }
    return Promise.resolve(false);
};

QLoader.get_cookie_val = function(cookiename) {
    let cookies = document.cookie.split(";");
    for(let i = 0; i < cookies.length; i++) {
        let c = cookies[i];
        let index = c.indexOf(cookiename);
        if (index >= 0) {
            return c.substring(index + cookiename.length + 5);
        }
    }
    return "";
};


QLoader.fetch_with_suffix_ajax = function (url, params, script_urls) {
    axios.post(url, params).then(function (resp) {
        let data = resp.data;
        if (data.code) {
            console.error("[loader]", url, params, data);
            return data;
        }
        QLoader.fetch_with_suffix(script_urls, data.data);
        return data;
    }.bind(this)).catch(function (ex) {
        console.error("[loader] fail", url, params, ex);
    }.bind(this));
};

