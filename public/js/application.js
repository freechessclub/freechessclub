var chat = new ReconnectingWebSocket(location.protocol.replace("http","ws") + "//" + location.host + "/ws");
var board = ChessBoard('board', 'start');
var connected = false;
var myHandle = "";
var tabs;

var msgType = {
  ctl: 0,
  chTell: 1,
  pTell: 2,
  gameMove: 3,
  unknown: 4
};

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
  delete tabs["content-"+tabContentId.substr(1)];
  $('#tabs a:last').tab('show');
  $(tabContentId).remove();
});

$(document).on('shown.bs.tab', 'a[data-toggle="tab"]', function (e) {
  var tab = $(e.target);
  tab.css('color', 'black');
});

function handleMsg(tab, data) {
  who = "";
  var tabheader = $("#" + $("ul#tabs a.active").attr("id"));
  if (data.hasOwnProperty('handle')) {
    var textclass = "";
    if (myHandle == data.handle) {
      textclass = " class=\"mine\"";
    }
    who = "<strong"+textclass+">"+$('<span/>').text(data.handle).html()+"</strong>: ";
    if (data.type == msgType.chTell) {
      tabheader = $("#"+data.channel);
    } else {
      tabheader = $("#"+data.handle);
    }
  }
  tab.append(who +
    $('<span/>').text(data.text).html()+"</br>");

  if (tabheader.hasClass('active')) {
    tab.scrollTop(tab[0].scrollHeight);
  } else {
    tabheader.css('color', 'red');
  }
}

chat.onmessage = function(message) {
  var data = JSON.parse(message.data);
  switch(data.type) {
    case msgType.ctl:
      if (connected == false && data.command == 1) {
        connected = true;
        myHandle = data.text;
        $("#chat-status").text("Connected as " + myHandle);
      }
      break;
    case msgType.chTell:
      var tab = tabs[data.channel];
      handleMsg(tab, data)
      break;
    case msgType.pTell:
      var tab;
      if (!tabs.hasOwnProperty(data.handle)) {
        $('<a class="flex-sm-fill text-sm-center nav-link" data-toggle="tab" href="#content-'+data.handle+'" id="'+data.handle+'" role="tab">'+data.handle+'<span class="btn btn-default btn-sm closeTab">×</span></a>').appendTo('#tabs');
        $('<div class="tab-pane chat-text" id="content-'+data.handle+'" role="tabpanel"></div>').appendTo('.tab-content');
        $(".chat-text").height($("#board").height()-115);
        tab = $("#content-"+data.handle);
        tabs[data.handle] = tab;
      } else {
        tab = tabs[data.handle];
      }
      handleMsg(tab, data);
      break;
    case msgType.gameMove:
      break;
    case msgType.unknown:
    default:
      var tab = $("ul#tabs a.active").attr("id");
      handleMsg(tabs[tab], data);
      break;
  }
};

chat.onclose = function(){
  connected = false;
  $("#chat-status").text("Disconnected");
};

$("#input-form").on("submit", function(event) {
  event.preventDefault();
  var tab = $("ul#tabs a.active").attr("id");
  var text;
  if (!$("#input-command").is(':checked')) {
    if ($("#input-text").val().charAt(0) != "@") {
      msg = $("#input-text").val();
      text = "t " + tab + " " + msg;
      handleMsg(tabs[tab], { type: msgType.chTell, channel: tab, handle: myHandle, text: msg });
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
  chat.send(JSON.stringify({ type: msgType.pTell, handle: myHandle, text: text }));
  $("#input-text").val("");
});

$(document).ready(function() {
  connected = false;
  $("#chat-status").text("Connecting...");
  $(".chat-text").height($("#board").height()-115);
  tabs = { "53": $("#content-53") };
});

$(window).resize(function() {
  board.resize();
  $(".chat-text").height($("#board").height()-115);
});
