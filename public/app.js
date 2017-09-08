const fs = require('fs');
const cheerio = require('cheerio');

$ = cheerio.load(fs.readFileSync('play.html'));
$('header').remove();
$('.footer').remove();
fs.writeFileSync('app.html', $.html());
