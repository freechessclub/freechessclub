const fs = require('fs');
const cheerio = require('cheerio');

$ = cheerio.load(fs.readFileSync('play.html'));
$('header').remove();
$('.footer').remove();
fs.writeFileSync('app.html', $.html());

$('head').append(`<meta http-equiv="Content-Security-Policy" content="default-src * 'self' data: gap: https://ssl.gstatic.com 'unsafe-eval' 'unsafe-inline'; style-src * 'self' 'unsafe-inline'; media-src * data:; img-src 'self' data: content:;">`)
$('head').append('<meta name="format-detection" content="telephone=no">')
$('head').append('<meta name="msapplication-tap-highlight" content="no">')

$('body').append('<script type="text/javascript" src="cordova.js"></script>')
fs.writeFileSync('www/index.html', $.html());

