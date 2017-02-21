var board = ChessBoard('board', 'start');
var chat = new ReconnectingWebSocket(location.protocol.replace("http","ws") + "//" + location.host + "/ws");
var connected = false;
var handle = "";

$(function () {
  $('[data-toggle="tooltip"]').tooltip()
})

$('#collapse-chat').on('hidden.bs.collapse', function () {
  $("#chat-toggle").text("+");
})
$('#collapse-chat').on('show.bs.collapse', function () {
  $("#chat-toggle").text(String.fromCharCode(8211));
})

chat.onmessage = function(message) {
  var data = JSON.parse(message.data);
  if (connected == false && data.handle != "") {
    connected = true;
    handle = data.handle;
    $("#chat-status").text("Connected as " + handle);
    if (data.text === "") {
      return;
    }
  }

  var textclass = "";
  if (handle == data.handle) {
    textclass = " class=\"mine\"";
  }

  $("#chat-text").append(
    "<strong"+textclass+">"+$('<span/>').text(data.handle).html()+"</strong>: "+
    $('<span/>').text(data.text).html()+"</br>");
  $("#chat-text").scrollTop($("#chat-text")[0].scrollHeight);
};

chat.onclose = function(){
  connected = false;
  $("#chat-status").text("Disconnected");
  this.chat = new WebSocket(chat.url);
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
  $("#chat-text").height($("#board").height()-55);
});

$(window).resize(function() {
  board.resize();
  $("#chat-text").height($("#board").height()-55);
});
