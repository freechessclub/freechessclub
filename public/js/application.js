var chat = new ReconnectingWebSocket(location.protocol.replace("http","ws") + "//" + location.host + "/ws");
var board = ChessBoard('board', 'start');
var connected = false;
var handle = "";
var tabs;

$(function () {
  $('[data-toggle="tooltip"]').tooltip()
})

$('#collapse-chat').on('hidden.bs.collapse', function () {
  $("#chat-toggle").text("+");
})
$('#collapse-chat').on('show.bs.collapse', function () {
  $("#chat-toggle").text(String.fromCharCode(8211));
})

jQuery(document.body).on('click', '.closeTab', function(event) {
  var tabContentId = $(this).parent().attr("href");
  $(this).parent().remove();
  delete tabs[tabContentId.substr(1)+"-content"];
  $('#tabs a:last').tab('show');
  $(tabContentId).remove();
});

$(document).on('shown.bs.tab', 'a[data-toggle="tab"]', function (e) {
  var tab = $(e.target);
  tab.css('color', 'black');
});

chat.onmessage = function(message) {
  var data = JSON.parse(message.data);
  if (connected == false && data.handle != "" && data.type == 0) {
    connected = true;
    handle = data.handle;
    $("#chat-status").text("Connected as " + handle);
    if (data.text == "") {
      return;
    }
  }

  var tab = tabs["ch53"];
  if (data.type == 2) {
    if (!tabs.hasOwnProperty(data.handle)) {
      $('<a class="flex-sm-fill text-sm-center nav-link" data-toggle="tab" href="#'+data.handle+'-content" id="'+data.handle+'" role="tab">'+data.handle+'<span class="btn btn-default btn-sm closeTab">×</span></a>').appendTo('#tabs');
      $('<div class="tab-pane chat-text" id="'+data.handle+'-content" role="tabpanel"></div>').appendTo('.tab-content');
      $(".chat-text").height($("#board").height()-115);
      tab = $("#"+data.handle+"-content");
      tabs[data.handle] = tab;
    } else {
      tab = tabs[data.handle];
    }
  }

  var textclass = "";
  if (handle == data.handle) {
    textclass = " class=\"mine\"";
  }

  if (data.type == 1 || data.type == 2 || data.type == 4) {
    tab.append(
      "<strong"+textclass+">"+$('<span/>').text(data.handle).html()+"</strong>: "+
      $('<span/>').text(data.text).html()+"</br>");
      tabheader = $("#"+data.handle);
      if (tabheader.hasClass('active')) {
        tab.scrollTop(tab[0].scrollHeight);
      } else {
        tabheader.css('color', 'red');
      }
  }
};

chat.onclose = function(){
  connected = false;
  $("#chat-status").text("Disconnected");
};

$("#input-form").on("submit", function(event) {
  event.preventDefault();
  var text;
  if (!$("#input-command").is(':checked')) {
    if ($("#input-text").val().charAt(0) != "@") {
      text = "t 53 " + $("#input-text").val()
    } else {
      text = $("#input-text").val().substr(1);
    }
  } else {
    if ($("#input-text").val().charAt(0) != "@") {
      text = $("#input-text").val();
    } else {
      text = $("#input-text").val().substr(1);
    }
  }
  chat.send(JSON.stringify({ handle: "", text: text }));
  $("#input-text").val("");
});

$(document).ready(function() {
  connected = false;
  $("#chat-status").text("Connecting...");
  $(".chat-text").height($("#board").height()-115);
  tabs = { ch53: $("#ch53-content") };
});

$(window).resize(function() {
  board.resize();
  $(".chat-text").height($("#board").height()-115);
});
