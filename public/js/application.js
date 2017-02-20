var board = ChessBoard('board', 'start');
var chat = new ReconnectingWebSocket(location.protocol.replace("http","ws") + "//" + location.host + "/ws");
var connected = false;
var handle = "";

chat.onmessage = function(message) {
  var data = JSON.parse(message.data);
  if (connected == false && data.handle != "") {
    connected = true;
    handle = data.handle;
    $('#chat-status').text("Connected as " + handle);
  }
  $("#chat-text").append(
    "<b>" + $('<span/>').text(data.handle).html() + "</b>: " +
    $('<span/>').text(data.text).html()+"</br>");
};

chat.onclose = function(){
  console.log('chat closed');
  this.chat = new WebSocket(chat.url);
};

$("#input-form").on("submit", function(event) {
  event.preventDefault();
  var text = "";
  if (!$("#input-command").is(':checked') &&
  ($("#input-text").val().charAt(0) != '+' ||
  $("#input-text").val().charAt(0) != '-')) {
    text += "t 53 "
  }
  text += $("#input-text").val();
  chat.send(JSON.stringify({ handle: "", text: text }));
  $("#input-text").val("");
});

$(document).ready(function() {
  connected = false;
  $('#chat-status').text("Connecting...");
});
