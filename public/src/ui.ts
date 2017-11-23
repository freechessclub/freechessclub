// Copyright 2017 Free Chess Club.

// enable tooltips
$(() => {
  $('[data-toggle="tooltip"]').tooltip();
});

if ($(window).width() < 767) {
  $('#collapse-chat').collapse('hide');
}

if ($(window).width() < 767) {
  $('#collapse-history').collapse('hide');
}

// color theme controls
$('#colortheme-default').on('click', (event) => {
  $('#colortheme').attr('href', 'www/css/themes/default.css');
});

$('#colortheme-green').on('click', (event) => {
  $('#colortheme').attr('href', 'www/css/themes/green.css');
});

$('#colortheme-yellow').on('click', (event) => {
  $('#colortheme').attr('href', 'www/css/themes/yellow.css');
});
