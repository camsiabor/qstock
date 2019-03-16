
var webpage = require('webpage');
var system = require('system');
if (system.args.length === 1) {
    console.log('Usage: loadspeed.js [some URL]');
    phantom.exit();
}

var url = system.args[1];
var page = webpage.create();
page.open(url, "GET", function (status) {
    if (status === "success") {
        var content = page.content;
        console.log(content);
    } else {
        console.log(status);
    }
    phantom.exit();
});