// Copyright 2017 Free Chess Club.

// enable tooltips
$(() => {
  $('[data-toggle="tooltip"]').tooltip();
});

// color theme controls
$('#colortheme-default').on('click', (event) => {
  $('#colortheme').attr('href', 'assets/css/themes/default.css');
});

$('#colortheme-green').on('click', (event) => {
  $('#colortheme').attr('href', 'assets/css/themes/green.css');
});

$('#colortheme-yellow').on('click', (event) => {
  $('#colortheme').attr('href', 'assets/css/themes/yellow.css');
});
